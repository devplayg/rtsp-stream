package streaming

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/devplayg/hippo"
	"github.com/devplayg/rtsp-stream/common"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Manager struct {
	server               *Server
	streams              map[int64]*Stream // Stream pool
	scheduler            *cron.Cron
	ctx                  context.Context
	cancel               context.CancelFunc
	watcherCheckInterval time.Duration
	sync.RWMutex
}

func NewManager(server *Server) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		server:               server,
		streams:              make(map[int64]*Stream), /* key: id(int64), value: &stream */
		ctx:                  ctx,
		cancel:               cancel,
		watcherCheckInterval: 20 * time.Second,
	}
}

func (m *Manager) init() error {
	if err := m.loadStreamsFromDatabase(); err != nil {
		return err
	}

	if err := m.initStreamDatabases(); err != nil {
		return err
	}

	if err := m.cleanStreamMetaFile(); err != nil {
		return err
	}

	if err := m.startScheduler(); err != nil {
		return err
	}

	return nil
}

func (m *Manager) initStreamDatabases() error {
	for id, _ := range m.streams {
		db, err := m.openStreamDB(id)
		if err != nil {
			return err
		}
		m.streams[id].db = db
	}
	return nil
}

func (m *Manager) start() error {
	if err := m.init(); err != nil {
		return err
	}

	go m.startStreamWatcher()

	return nil
}

func (m *Manager) cleanStreamMetaFile() error {
	m.Lock()
	defer m.Unlock()

	dir := filepath.Join(m.server.config.Storage.LiveDir)
	for id, stream := range m.streams {
		path := filepath.ToSlash(filepath.Join(dir, strconv.FormatInt(stream.Id, 10), stream.ProtocolInfo.MetaFileName))
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			//err := os.Remove(path)
			log.WithFields(log.Fields{
				"err":  err,
				"file": filepath.Base(path),
			}).Debugf("[manager] cleaned meta file of stream-%d", id)
		}
	}

	return nil
}

func (m *Manager) loadStreamsFromDatabase() error {
	m.Lock()
	defer m.Unlock()
	return common.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(common.StreamBucket)
		return b.ForEach(func(k, v []byte) error {
			var stream Stream
			err := json.Unmarshal(v, &stream)
			if err != nil {
				log.Error(err)
				return nil
			}
			stream.Status = common.Stopped
			m.streams[stream.Id] = &stream
			return nil
		})
	})

	// wondory
	// fetch and unmarshal
	// lock
	// assign
	// unlock
}

func (m *Manager) getStreams() []*Stream {
	streams := make([]*Stream, 0)
	m.RLock()
	defer m.RUnlock()
	for _, stream := range m.streams {
		streams = append(streams, stream)
	}

	return streams
}

func (m *Manager) getStreamById(id int64) *Stream {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.streams[id]; !ok {
		return nil
	}
	stream := m.streams[id]
	if stream.cmd != nil && stream.cmd.Process != nil {
		stream.Pid = stream.cmd.Process.Pid
	}

	return stream
}

func (m *Manager) addStream(stream *Stream) error {
	if err := m.isValidStream(stream); err != nil {
		return err
	}

	if err := m.issueStream(stream); err != nil {
		return err
	}

	db, err := m.openStreamDB(stream.Id)
	if err != nil {
		return err
	}
	stream.db = db
	log.WithFields(log.Fields{
		"stream_id": stream.Id,
		"uri":       stream.Uri,
	}).Debugf("[manager] issued new stream")

	return nil
}

func (m *Manager) isValidStream(stream *Stream) error {
	if len(stream.Uri) < 1 {
		return errors.New("empty stream url")
	}
	stream.UrlHash = GetHashString(stream.Uri)

	if !(stream.Protocol == common.HLS || stream.Protocol == common.WEBM) {
		return errors.New("unknown stream protocol: " + strconv.Itoa(stream.Protocol))
	}
	stream.ProtocolInfo = common.NewProtocolInfo(stream.Protocol)

	return nil
}

func (m *Manager) issueStream(input *Stream) error {
	id, err := IssueStreamId()
	if err != nil {
		return err
	}
	input.Id = id
	m.streams[input.Id] = input

	return SaveStreamInDB(input)
}

func (m *Manager) updateStream(stream *Stream) error {
	if err := m.isValidStream(stream); err != nil {
		return err
	}

	m.Lock()
	m.streams[stream.Id] = stream
	m.Unlock()

	return SaveStreamInDB(stream)
}

func (m *Manager) deleteStream(id int64) error {
	if err := m.stopStreaming(id); err != nil {
		return err
	}
	if err := m.closeStreamDB(id); err != nil {
		return err
	}

	if err := os.Remove(filepath.Join(m.server.dbDir, m.streams[id].getDbFileName())); err != nil {
		return err
	}

	m.Lock()
	delete(m.streams, id)
	m.Unlock()

	err := DeleteStreamInDB(id)
	log.WithFields(log.Fields{}).Infof("[manager] deleted stream-%d", id)

	return err
}

func (m *Manager) cleanStreamDir(stream *Stream) error {
	files, err := ioutil.ReadDir(stream.liveDir)
	if err != nil {
		return err
	}
	t := time.Now().In(common.Loc)
	for _, f := range files {
		if f.ModTime().In(common.Loc).Format(common.DateFormat) == t.Format(common.DateFormat) {
			continue
		}
		if err := os.Remove(filepath.Join(stream.liveDir, f.Name())); err != nil {
			log.Error(err)
			continue
		}
	}

	return nil
}

func (m *Manager) removeStreamDir(stream *Stream) error {
	err := os.RemoveAll(stream.liveDir)
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) createStreamDir(stream *Stream) error {
	stream.liveDir = filepath.ToSlash(filepath.Join(m.server.config.Storage.LiveDir, strconv.FormatInt(stream.Id, 10)))
	if err := hippo.EnsureDir(stream.liveDir); err != nil {
		return err
	}

	return nil
}

func (m *Manager) changeStreamStatusToStart(id int64) (*Stream, error) {
	// Check stream status
	m.Lock()
	defer m.Unlock()

	stream := m.streams[id]
	if stream == nil {
		return nil, common.ErrorStreamNotFound
	}
	if stream.Status == common.Started {
		return nil, errors.New(fmt.Sprintf("[manager] stream-%d has been already started", id))
	}
	if stream.Status == common.Starting {
		return nil, errors.New(fmt.Sprintf("[manager] stream-%d is already starting now", id))
	}
	if stream.Status == common.Stopping {
		return nil, errors.New(fmt.Sprintf("[manager] stream-%d is already stopping now", id))
	}
	stream.Status = common.Starting
	return stream, nil
}

func (m *Manager) startStreaming(id int64, from string) error {
	// Who sent?
	log.WithFields(log.Fields{
		"from": from,
	}).Infof("[manager] received to start stream-%d", id)

	stream, err := m.changeStreamStatusToStart(id)
	if err != nil {
		return err
	}

	if err := m.createStreamDir(stream); err != nil {
		stream.Status = common.Failed
		return err
	}

	if err := m.cleanStreamDir(stream); err != nil {
		stream.Status = common.Failed
		return err
	}

	go func() {
		count, err := stream.start()
		if err != nil {
			log.WithFields(log.Fields{
				"id": id,
			}).Errorf("[manager] failed to start stream-%d: %s", id, err)
			stream.Status = common.Failed
			return
		}
		log.WithFields(log.Fields{
			"id":        id,
			"url":       stream.Uri,
			"waitCount": count,
			"pid":       GetStreamPid(stream),
		}).Infof("[manager] stream-%d has been started", id)
		stream.Status = common.Started
	}()

	return nil
}

func (m *Manager) stopStreaming(id int64) error {
	log.WithFields(log.Fields{}).Infof("[manager] received to stop stream-%d", id)

	m.Lock()
	defer m.Unlock()

	stream := m.streams[id]
	if stream == nil {
		return common.ErrorStreamNotFound
	}
	if stream.Status == common.Stopped {
		return nil
	}

	if stream.Status == common.Stopping {
		return errors.New(fmt.Sprintf("[manager] stream-%d is already stopping now", id))
	}
	if stream.Status == common.Starting {
		return errors.New(fmt.Sprintf("[manager] stream-%d is already starting now", id))
	}
	stream.Status = common.Stopping
	stream.stop()

	return nil
}

func (m *Manager) Stop() error {
	m.cancel()
	for id, _ := range m.streams {
		m.stopStreaming(id)
		if err := m.streams[id].db.Close(); err != nil {
			log.Error(err)
		}
	}
	log.Debug("[manager] all streams have been stopped")

	return nil
}
func (m *Manager) startStreamWatcher() {
	log.WithFields(log.Fields{
		"interval": fmt.Sprintf("%3.1fsec", m.watcherCheckInterval.Seconds()),
	}).Debug("[manager] watcher has been started")
	for {
		for id, stream := range m.streams {
			if !stream.Enabled {
				continue
			}

			// just in case
			if stream.Status == common.Started && !stream.IsActive() {
				log.WithFields(log.Fields{}).Errorf("###[stream-%d]### status is 'started' but stream wasn't alive.", stream.Id)
				// stream.stop()
				if err := m.stopStreaming(id); err != nil {
					log.Error(err)
				}
			}
			if stream.Status != common.Started && stream.IsActive() {
				log.WithFields(log.Fields{}).Errorf("###[stream-%d]### status is not 'started' but it's alive!!!", stream.Id)
			}

			if stream.IsActive() {
				continue
			}

			log.WithFields(log.Fields{}).Infof("[watcher] since stream-%d is not running, start it", id)
			if err := m.startStreaming(id, "watcher"); err != nil {
				log.Error(err)
				continue
			}
		}

		select {
		case <-time.After(m.watcherCheckInterval):
		case <-m.ctx.Done():
			log.Debug("[manager] stream watcher has been stopped")
			return
		}
	}
}

func (m *Manager) getM3u8(id int64, date string) (string, error) {
	stream := m.getStreamById(id)
	if stream == nil {
		return "", common.ErrorStreamNotFound
	}

	segments, err := stream.getM3u8Segments(date)
	if err != nil {
		return "", err
	}
	tags := stream.makeM3u8Tags(segments)
	return tags, nil
}

func (m *Manager) openStreamDB(id int64) (*bolt.DB, error) {
	path := filepath.Join(m.server.dbDir, "stream-"+strconv.FormatInt(id, 10)+".db")
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (m *Manager) closeStreamDB(id int64) error {
	return m.streams[id].db.Close()
}

func (m *Manager) getVideoRecords() ([]map[string]string, error) {
	bucketNames, err := common.GetDbBucketList(common.DB, common.VideoBucketPrefix)
	if err != nil {
		return nil, err
	}
	dayRecordMap, err := m.getPrevVideoRecords(bucketNames)
	dayRecordMap[common.LiveBucketName] = m.getLiveVideoStatus(bucketNames)
	return common.SortDayRecord(dayRecordMap), err
}

func (m *Manager) getLiveVideoStatus(bucketNames []string) map[string]string {
	liveMap := common.CreateDefaultDayRecord("live", bucketNames)
	for _, bn := range bucketNames {
		streamId, err := strconv.ParseInt(strings.TrimPrefix(bn, common.VideoBucketPrefix), 10, 16)
		if err != nil {
			log.WithFields(log.Fields{
				"bucketName": bn,
			}).Error(err)
			continue
		}

		stream := m.getStreamById(streamId)
		if stream == nil {
			continue
		}

		if !stream.IsActive() {
			continue
		}

		liveMap[bn] = "active"
	}

	return liveMap
}

func (m *Manager) getPrevVideoRecords(bucketNames []string) (common.DayRecordMap, error) {
	dayRecordMap := make(common.DayRecordMap)
	err := common.DB.View(func(tx *bolt.Tx) error {
		for _, bn := range bucketNames {
			b := tx.Bucket([]byte(bn))
			b.ForEach(func(key, _ []byte) error {
				date := string(key)
				if _, ok := dayRecordMap[date]; !ok {
					dayRecordMap[date] = common.CreateDefaultDayRecord(date, bucketNames)
				}
				dayRecordMap[date][bn] = "1"
				return nil
			})
		}
		return nil
	})
	return dayRecordMap, err
}
