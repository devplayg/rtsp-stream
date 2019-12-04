package streaming

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/devplayg/hippo"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type Manager struct {
	server               *Server
	streams              map[int64]*Stream // Stream pool
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
		watcherCheckInterval: 5 * time.Second,
	}
}

func (m *Manager) start() error {
	if err := m.loadStreamsFromDatabase(); err != nil {
		return err
	}
	m.startStreamWatcher()

	return nil
}

func (m *Manager) loadStreamsFromDatabase() error {
	m.Lock()
	defer m.Unlock()
	return DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(StreamBucket)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var stream Stream
			err := json.Unmarshal(v, &stream)
			if err != nil {
				log.Error(err)
				continue
			}
			stream.Status = Stopped
			m.streams[stream.Id] = &stream
		}
		return nil
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
	stream := m.streams[id]
	if stream.cmd != nil && stream.cmd.Process != nil {
		stream.Pid = stream.cmd.Process.Pid
	}
	m.Unlock()
	return stream
}

func (m *Manager) addStream(stream *Stream) error {
	// Check if the stream is valid
	if err := m.isValidStream(stream); err != nil {
		return err
	}

	err := m.issueStream(stream)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"stream_id": stream.Id,
		"uri":       stream.Uri,
		"err":       err,
	}).Debugf("issue new stream")

	return err
}

func (m *Manager) isValidStream(stream *Stream) error {
	if len(stream.Uri) < 1 {
		return errors.New("empty stream url")
	}
	stream.UrlHash = GetHashString(stream.Uri)

	if !(stream.Protocol == HLS || stream.Protocol == WEBM) {
		return errors.New("unknown stream protocol: " + strconv.Itoa(stream.Protocol))
	}
	stream.ProtocolInfo = NewProtocolInfo(stream.Protocol)

	return nil
}

func (m *Manager) issueStream(input *Stream) error {
	var maxStreamId int64
	m.Lock()
	for id, stream := range m.streams {
		if input.UrlHash == stream.UrlHash {
			return errors.New("duplicated stream uri:" + input.Uri)
		}
		if maxStreamId < id {
			maxStreamId = id
		}
	}
	maxStreamId++ // issue new stream ID
	input.Id = maxStreamId
	m.streams[maxStreamId] = input
	m.Unlock()
	return SaveStream(input)
}

func (m *Manager) updateStream(stream *Stream) error {
	if err := m.isValidStream(stream); err != nil {
		return err
	}

	m.Lock()
	m.streams[stream.Id] = stream
	m.Unlock()

	return SaveStream(stream)
}

func (m *Manager) deleteStream(id int64) error {
	m.Lock()
	delete(m.streams, id)
	m.Unlock()

	return DeleteStream(id)
}

func (m *Manager) cleanStreamDir(stream *Stream) error {
	// Remove all files but created today in live directory

	files, err := ioutil.ReadDir(stream.liveDir)
	if err != nil {
		return err
	}
	t := time.Now().In(Loc)
	for _, f := range files {
		if f.ModTime().In(Loc).Format(DateFormat) == t.Format(DateFormat) {
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
	stream.liveDir = filepath.ToSlash(filepath.Join(m.server.liveDir, strconv.FormatInt(stream.Id, 10)))
	if err := hippo.EnsureDir(stream.liveDir); err != nil {
		return err
	}

	return nil
}

func (m *Manager) startStreaming(id int64, from string) error {
	// Who sent?
	log.WithFields(log.Fields{
		"from": from,
	}).Infof("[manager] received to start stream-%d", id)

	// Check stream status
	m.Lock()
	stream := m.streams[id]
	if stream == nil {
		m.Unlock()
		return ErrorStreamNotFound
	}

	if stream.Status == Started {
		log.Warnf("[manager] stream-%d has been already started", id)
		m.Unlock()
		return nil
	}
	if stream.Status == Starting {
		log.Warnf("[manager] stream-%d is already starting now", id)
		m.Unlock()
		return nil
	}
	if stream.Status == Stopping {
		log.Warnf("[manager] stream-%d is already stopping now", id)
		m.Unlock()
		return nil
	}
	stream.Status = Starting
	m.Unlock()

	go func() {
		if err := m.createStreamDir(stream); err != nil {
			log.Error(err)
			stream.Status = Failed
			return
		}

		if err := m.cleanStreamDir(stream); err != nil {
			log.Warn("failed to clean streaming directories:", err)
			stream.Status = Failed
			return
		}

		count, err := stream.start()
		if err != nil {
			log.WithFields(log.Fields{
				"id": id,
			}).Errorf("[manager] failed to start stream-%d: %s", id, err)
			m.RLock()
			stream.Status = Failed
			m.Unlock()
			return
		}
		log.WithFields(log.Fields{
			"id":        id,
			"url":       stream.Uri,
			"waitCount": count,
			"pid":       GetStreamPid(stream),
		}).Infof("[manager] stream-%d has been started", id)
		stream.Status = Started
	}()

	return nil
}

func (m *Manager) stopStreaming(id int64) error {
	log.WithFields(log.Fields{}).Infof("[manager] received to stop stream-%d", id)

	m.Lock()
	defer m.Unlock()

	stream := m.streams[id]
	if stream == nil {
		return ErrorStreamNotFound
	}
	if stream.Status == Stopped {
		log.Warnf("[manager] stream-%d has been already stopped", id)
		return nil
	}
	if stream.Status == Stopping {
		log.Warnf("[manager] stream-%d is already stopping now", id)
		return nil
	}
	if stream.Status == Starting {
		log.Warnf("[manager] stream-%d is already starting now", id)
		return nil
	}
	stream.Status = Stopping
	stream.stop()

	return nil
}

func (m *Manager) Stop() error {
	m.cancel()
	for _, stream := range m.streams {
		//if stream.IsActive() {
		stream.stop()
		//}
	}
	log.Debug("[manager] all streams have been stopped")

	return nil
}
func (m *Manager) startStreamWatcher() {
	go func() {
		log.WithFields(log.Fields{
			"interval": fmt.Sprintf("%3.1fsec", m.watcherCheckInterval.Seconds()),
		}).Debug("[manager] stream watcher hash been started")
		for {
			m.checkStreams()

			select {
			case <-time.After(m.watcherCheckInterval):
			case <-m.ctx.Done():
				log.Debug("[manager] stream watcher has been stopped")
				return
			}
		}
	}()
}

func (m *Manager) getM3u8(id int64) (string, error) {
	stream := m.getStreamById(id)
	if stream == nil {
		return "", ErrorStreamNotFound
	}

	segs := stream.getM3u8Segments("")
	tags := stream.makeM3u8Tags(segs)

	return tags, nil

}

func (m *Manager) checkStreams() {
	for id, stream := range m.streams {
		if !stream.Enabled {
			continue
		}
		if stream.IsActive() {
			continue
		}

		log.WithFields(log.Fields{}).Errorf("[watcher] since stream-%d is not running, start it", id)
		if err := m.startStreaming(id, "watcher"); err != nil {
			log.Error(err)
		}
	}

}
