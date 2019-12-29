package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/davecgh/go-spew/spew"
	"github.com/devplayg/hippo"
	"github.com/devplayg/rtsp-stream/common"
	"github.com/devplayg/rtsp-stream/streaming"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Manager struct {
	server               *Server
	streams              map[int64]*streaming.Stream // Stream pool
	scheduler            *cron.Cron
	ctx                  context.Context
	cancel               context.CancelFunc
	watcherCheckInterval time.Duration
	onArchiving          bool
	sync.RWMutex
}

func NewManager(server *Server) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		server:               server,
		streams:              make(map[int64]*streaming.Stream), /* key: id(int64), value: &stream */
		ctx:                  ctx,
		cancel:               cancel,
		watcherCheckInterval: 15 * time.Second,
	}
}

func (m *Manager) init() error {
	if err := m.loadStreamsFromDatabase(); err != nil {
		return err
	}

	if err := m.initStreamDatabases(); err != nil {
		return err
	}

	//if err := m.cleanStreamMetaFile(); err != nil {
	//	return err
	//}

	if err := m.startScheduler(); err != nil {
		return err
	}

	//
	//s2 := rand.NewSource(42)
	//r2 := rand.New(s2)
	//for _, id := range m.getStreamIdList() {
	//	for i:=0; i<29; i++ {
	//		n := r2.Intn(29) + 1
	//		// fmt.Printf("201912%02d\n", n)
	//		m.writeVideoArchivingHistory(id, fmt.Sprintf("201912%02d\n", n))
	//	}
	//}

	return nil
}

func (m *Manager) getLastArchivingDate(t time.Time) (string, error) {
	val, err := GetValueFromDbBucket(common.ConfigBucket, common.LastArchivingDateKey)
	if err != nil {
		return "", err
	}
	if val == nil {
		return t.Add(7 * -24 * time.Hour).Format(common.DateFormat), nil
	}
	return string(val), nil
}

func (m *Manager) checkOldLiveVideoFiles() error {
	t := time.Now().In(common.Loc)
	lastArchivingDate, err := m.getLastArchivingDate(t)
	if err != nil {
		return err
	}
	expectedDate := t.Add(-24 * time.Hour).Format(common.DateFormat) // yesterday
	log.WithFields(log.Fields{
		"last":     lastArchivingDate,
		"expected": expectedDate,
	}).Debug("[manager] checking last archiving date")

	if lastArchivingDate == expectedDate {
		PutDataIntoDbBucket(common.ConfigBucket, common.LastArchivingDateKey, []byte(expectedDate))
		return nil
	}

	from, err := time.ParseInLocation(common.DateFormat, lastArchivingDate, common.Loc)
	if err != nil {
		return err
	}
	to, err := time.ParseInLocation(common.DateFormat, expectedDate, common.Loc)
	if err != nil {
		return err
	}
	if from.After(to) {
		return errors.New("invalid system time on scheduler")
	}

	d := from
	for d.Before(to) || d.Equal(to) {
		log.WithFields(log.Fields{
			"targetDate": d.Format(common.DateFormat),
		}).Debug("[manager] handling missed archiving task")
		if err := m.startToArchiveVideos(d.Format(common.DateFormat)); err != nil {
			log.Error(err)
		}
		d = d.Add(24 * time.Hour)
	}

	PutDataIntoDbBucket(common.ConfigBucket, common.LastArchivingDateKey, []byte(expectedDate))

	//err := common.DB.View(func(tx *bolt.Tx) error {
	//    b := tx.Bucket(common.ConfigBucket)
	//    value := b.Get(common.LastArchivingDateKey)
	//    if value == nil {
	//        lastArchivingDate = t.Add(7 * -24*time.Hour).Format(common.DateFormat)
	//        return nil
	//    }
	//    lastArchivingDate = string(value)
	//    return nil
	//})
	//if err != nil {
	//    return err
	//}

	return nil
}

func (m *Manager) initStreamDatabases() error {
	for id, _ := range m.streams {
		db, err := m.openStreamDB(id)
		if err != nil {
			return err
		}
		m.streams[id].DB = db
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

//
//func (m *Manager) cleanStreamMetaFile() error {
//	m.Lock()
//	defer m.Unlock()
//
//	dir := filepath.Join(m.server.config.Storage.LiveDir)
//	for id, stream := range m.streams {
//		path := filepath.ToSlash(filepath.Join(dir, strconv.FormatInt(stream.Id, 10), stream.ProtocolInfo.MetaFileName))
//		if _, err := os.Stat(path); !os.IsNotExist(err) {
//			err := os.Remove(path)
//			log.WithFields(log.Fields{
//				"err":  err,
//				"file": filepath.Base(path),
//			}).Debugf("[manager] cleaned meta file of stream-%d", id)
//		}
//	}
//
//	return nil
//}

func (m *Manager) loadStreamsFromDatabase() error {
	m.Lock()
	defer m.Unlock()
	return db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(common.StreamBucket)
		b.ForEach(func(k, v []byte) error {
			var stream streaming.Stream
			err := json.Unmarshal(v, &stream)
			if err != nil {
				spew.Dump(k)
				log.Error(err)
				return nil
			}
			stream.Status = common.Stopped
			m.streams[stream.Id] = &stream
			log.WithFields(log.Fields{
				"url":       stream.Uri,
				"recording": stream.Recording,
				"enabled":   stream.Enabled,
			}).Debugf("[manager] 'stream-%d' has been loaded", stream.Id)
			return nil
		})

		//b = tx.Bucket(common.ConfigBucket)
		//lastRecordingDate := b.Get(common.LastRecordingDateKey)
		log.WithFields(log.Fields{
			//"lastRecordingDate": string(lastRecordingDate),
		}).Debugf("[manager] %d stream(s) has been loaded", len(m.streams))
		return nil
	})

	//return err

	// wondory
	// fetch and unmarshal
	// lock
	// assign
	// unlock
}

func (m *Manager) getStreams() []*streaming.Stream {
	streams := make([]*streaming.Stream, 0)
	m.RLock()
	for _, stream := range m.streams {
		streams = append(streams, stream)
	}
	m.RUnlock()
	sort.Slice(streams, func(i, j int) bool {
		if streams[i].Name == streams[j].Name {
			return streams[i].Id < streams[j].Id
		}
		return streams[i].Name < streams[j].Name
	})

	return streams
}

func (m *Manager) getSimpleStreams() []*streaming.SimpleStream {
	streams := make([]*streaming.SimpleStream, 0)
	m.RLock()
	for _, stream := range m.streams {
		streams = append(streams, stream.Simplify())
	}
	m.RUnlock()
	sort.Slice(streams, func(i, j int) bool {
		if streams[i].Name == streams[j].Name {
			return streams[i].Id < streams[j].Id
		}
		return streams[i].Name < streams[j].Name
	})

	return streams
}

func (m *Manager) getStreamById(id int64) *streaming.Stream {
	m.Lock()
	if _, ok := m.streams[id]; !ok {
		return nil
	}
	stream := m.streams[id]
	m.Unlock()

	if stream.Cmd != nil && stream.Cmd.Process != nil {
		stream.Pid = stream.Cmd.Process.Pid
	}

	return stream
}

func (m *Manager) addStream(stream *streaming.Stream) error {
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
	stream.DB = db
	log.WithFields(log.Fields{
		"stream_id": stream.Id,
		"uri":       stream.Uri,
	}).Debugf("[manager] issued new stream")

	return nil
}

func (m *Manager) isValidStream(stream *streaming.Stream) error {
	if len(stream.Uri) < 1 {
		return errors.New("empty stream url")
	}
	if _, err := url.Parse(stream.Uri); err != nil {
		return common.ErrorInvalidUri
	}
	stream.UrlHash = common.GetHashString(stream.Uri)

	//if !(stream.Protocol == common.HLS || stream.Protocol == common.WEBM) {
	//	return errors.New("unknown stream protocol: " + strconv.Itoa(stream.Protocol))
	//}
	stream.SetProtocol(common.HLS)
	//stream.Protocol = common.HLS
	//stream.ProtocolInfo = common.NewProtocolInfo(common.HLS)

	return nil
}

func (m *Manager) issueStream(input *streaming.Stream) error {
	id, err := IssueStreamId()
	if err != nil {
		return err
	}
	input.Id = id
	input.Created = time.Now().Unix()
	input.Updated = input.Created
	input.SetProtocol(common.HLS)
	m.streams[input.Id] = input

	return m.saveStream(input)

}

func (m *Manager) saveStream(stream *streaming.Stream) error {
	b, err := json.Marshal(stream)
	if err != nil {
		return err
	}

	return PutDataIntoDbBucket(common.StreamBucket, common.CreateStreamKey(stream.Id), b)
}

func (m *Manager) updateStream(input *streaming.Stream) error {
	if err := m.isValidStream(input); err != nil {
		return err
	}

	//m.Lock()
	//m.streams[stream.Id] = stream
	//m.Unlock()
	stream := m.getStreamById(input.Id)
	if stream == nil {
		return common.ErrorInvalidStream
	}

	m.RLock()
	defer m.RUnlock()
	stream.Name = input.Name
	stream.Uri = input.Uri
	stream.Enabled = input.Enabled
	stream.Recording = input.Recording
	stream.Username = input.Username
	stream.Password = input.Password
	stream.ProtocolInfo = input.ProtocolInfo
	stream.UrlHash = input.UrlHash
	stream.Updated = time.Now().Unix()

	return m.saveStream(stream)
	// return SaveStreamInDB(stream)
}

func (m *Manager) deleteStream(id int64, from string) error {
	if err := m.stopStreaming(id, from); err != nil {
		return err
	}
	if err := m.closeStreamDB(id); err != nil {
		return err
	}

	if err := os.Remove(filepath.Join(m.server.dbDir, m.streams[id].GetDBFileName())); err != nil {
		return err
	}

	m.Lock()
	delete(m.streams, id)
	m.Unlock()

	err := DeleteStreamInDB(id)
	log.WithFields(log.Fields{}).Infof("[manager] deleted stream-%d", id)

	return err
}

func (m *Manager) cleanStreamDir(stream *streaming.Stream) error {
	files, err := ioutil.ReadDir(stream.GetLiveDir())
	if err != nil {
		return err
	}

	//t := time.Now().In(common.Loc)
	for _, f := range files {
		if f.IsDir() {
			continue
		}

		if !f.Mode().IsRegular() {
			continue
		}

		if !strings.HasSuffix(f.Name(), common.VideoFileExt) {
			continue
		}
		//if f.ModTime().In(common.Loc).Format(common.DateFormat) == t.Format(common.DateFormat) {
		//	continue
		//}
		if f.Size() < 1 {
			if err := os.Remove(filepath.Join(stream.GetLiveDir(), f.Name())); err != nil {
				log.Error(err)
				continue
			}
		}
	}

	return nil
}

//
//func (m *Manager) removeStreamDir(stream *Stream) error {
//	err := os.RemoveAll(stream.liveDir)
//	if err != nil {
//		return err
//	}
//	return nil
//}

func (m *Manager) createStreamDir(stream *streaming.Stream) error {
	stream.SetLiveDir(filepath.ToSlash(filepath.Join(m.server.config.Storage.LiveDir, strconv.FormatInt(stream.Id, 10))))
	if err := hippo.EnsureDir(stream.GetLiveDir()); err != nil {
		return err
	}

	return nil
}

func (m *Manager) changeStreamStatusToStart(id int64) (*streaming.Stream, error) {
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
		count, err := stream.Start()
		if err != nil {
			log.WithFields(log.Fields{
				"id": id,
			}).Errorf("[manager] failed to start stream-%d: %s", id, err)
			//stream.Status = common.Failed
			return
		}
		log.WithFields(log.Fields{
			"id":        id,
			"url":       stream.Uri,
			"waitCount": count,
			"pid":       streaming.GetStreamPid(stream),
		}).Infof("[manager] stream-%d has been started", id)
		//stream.Status = common.Started
	}()

	return nil
}

func (m *Manager) stopStreaming(id int64, from string) error {
	log.WithFields(log.Fields{
		"from": from,
	}).Infof("[manager] received to stop stream-%d", id)

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
	stream.Stop()

	return nil
}

func (m *Manager) Stop() error {
	m.cancel()
	for id, _ := range m.streams {
		m.stopStreaming(id, "manager")
		if m.streams[id].DB == nil {
			continue
		}
		if err := m.streams[id].DB.Close(); err != nil {
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
			active, lastStreamUpdated, diff := stream.GetStatus()
			stream.LastStreamUpdated = lastStreamUpdated
			if !stream.Enabled {
				continue
			}

			// just in case (if you restart immediately after stopping)
			if !active && stream.Status == common.Started {
				log.WithFields(log.Fields{
					"lastStreamUpdated": lastStreamUpdated.Format(time.RFC3339),
					"diff":              diff,
				}).Errorf("###[stream-%d]### status is 'started' but stream wasn't alive.", stream.Id)
				if err := m.stopStreaming(id, "watcher"); err != nil {
					log.Error(err)
				}
			}
			// just in case
			if active && stream.Status != common.Started {
				log.WithFields(log.Fields{
					"status": stream.Status,
				}).Errorf("###[stream-%d]### status is not 'started' but it's alive!!!", stream.Id)
			}

			if active {
				continue
			}

			if time.Since(stream.LastAttemptTime) < 10*time.Second {
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

func (m *Manager) getM3u8(id int64, date string) (string, error) { // wondory
	stream := m.getStreamById(id)
	if stream == nil {
		return "", common.ErrorStreamNotFound
	}

	tags, err := stream.GetM3u8Tags(date)
	return tags, err
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
	if m.streams[id].DB == nil {
		return errors.New("no database")
	}
	return m.streams[id].DB.Close()
}

func (m *Manager) getVideoRecords() ([]map[string]interface{}, error) {
	videoMap, dateMap, err := common.GetVideoRecordHistory(db)
	if err != nil {
		return nil, err
	}

	videoNames := make([]string, 0, len(videoMap)) // video-1, video-2...
	for k, _ := range videoMap {
		videoNames = append(videoNames, k)
	}
	dates := make([]string, 0, len(dateMap))
	for d, _ := range dateMap {
		dates = append(dates, d)
	}
	sort.Slice(dates, func(i, j int) bool {
		return dates[i] > dates[j]
	})

	videoArr := make([]map[string]interface{}, 0)
	for _, d := range dates {
		m := common.CreateDefaultDayRecord(d, videoNames)
		videoArr = append(videoArr, m)
		for _, videoName := range videoNames {
			if _, ok := videoMap[videoName][d]; !ok {
				continue
			}
			m[videoName] = 1
		}
	}

	return videoArr, err
}

//func (m *Manager) getVideoRecords() (map[string]interface{}, error) {
//	streams := m.getSimpleStreams()
//	if len(streams) < 1 {
//		return nil, nil
//	}
//	t := time.Now().In(common.Loc)
//	lastArchivingDateKey, _ := GetValueFromDbBucket(common.ConfigBucket, common.LastArchivingDateKey)
//	result := map[string]interface{}{
//		"date":                 t.Format(common.DateFormat),
//		"lastArchivingDateKey": string(lastArchivingDateKey),
//		"streams":              streams,
//		"videos":               nil,
//	}
//
//	bucketNames := m.convertStreamsToBucketNames(streams)
//	dayRecordMap, err := m.getPrevVideoRecords(bucketNames)
//	if err != nil {
//		return nil, err
//	}
//	dayRecordMap[common.LiveBucketName] = m.getLiveVideoStatus(bucketNames, t.Format(common.DateFormat))
//	result["videos"] = common.SortDayRecord(dayRecordMap)
//	return result, err
//}

//func (m *Manager) getLiveVideoStatus(bucketNames []string, date string) map[string]string {
//	liveMap := common.CreateDefaultDayRecord("live", bucketNames)
//	for _, bn := range bucketNames {
//		streamId, err := strconv.ParseInt(strings.TrimPrefix(bn, common.VideoBucketPrefix), 10, 64)
//		if err != nil {
//			log.WithFields(log.Fields{
//				"bucketName": bn,
//			}).Error(err)
//			continue
//		}
//
//		stream := m.getStreamById(streamId)
//		if stream == nil {
//			continue
//		}
//
//		if !stream.IsActive() {
//			continue
//		}
//
//		liveMap[bn] = "1"
//		if stream.M3u8BucketExists(date) {
//			liveMap[bn] += ",1"
//		}
//	}
//
//	return liveMap
//}
//
//func (m *Manager) getPrevVideoRecords() (common.DayRecordMap, error) {
//	dayRecordMap := make(common.DayRecordMap)
//	err := db.View(func(tx *bolt.Tx) error {
//		//for _, bn := range bucketNames {
//		//	b := tx.Bucket([]byte(bn))
//		//	if b == nil {
//		//		return nil
//		//	}
//		//	b.ForEach(func(key, _ []byte) error {
//		//		date := string(key)
//		//		if _, ok := dayRecordMap[date]; !ok {
//		//			dayRecordMap[date] = common.CreateDefaultDayRecord(date, bucketNames)
//		//		}
//		//		dayRecordMap[date][bn] = "1"
//		//		return nil
//		//	})
//		//}
//		return nil
//	})
//	return dayRecordMap, err
//}

//func (m *Manager) convertStreamsToBucketNames(streams []*streaming.SimpleStream) []string {
//	if len(streams) < 1 {
//		return nil
//	}
//	bucketNames := make([]string, 0)
//	for _, s := range streams {
//		bucketName := common.VideoBucketPrefix + strconv.FormatInt(s.Id, 10)
//		bucketNames = append(bucketNames, bucketName)
//	}
//	return bucketNames
//}

func (m *Manager) getLiveData() map[string]interface{} {
	streams := m.getStreams()

	return map[string]interface{}{
		"streams": streams,
	}

}

//func (m *Manager) toggleStreamEnabled()
