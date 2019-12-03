package streaming

import (
	"context"
	"encoding/json"
	"errors"
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
		watcherCheckInterval: 20 * time.Second,
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

func (m *Manager) canMakeStreamStatus(id int64, want int) (*Stream, error) {
	m.Lock()
	stream := m.streams[id]
	m.Unlock()

	if stream == nil {
		return nil, ErrorStreamNotFound
	}
	if want == Stopped {
		if stream.Status == Stopped {
			return nil, errors.New("stream is already stopped")
		}
		if stream.Status == Stopping {
			return nil, errors.New("stream is stopping now")
		}

		if stream.Status == Starting {
			return nil, errors.New("stream is about to start now")
		}

		//stream.Status = status
		return stream, nil
	}

	if stream.Status == Stopping {
		return nil, errors.New("stream is stopping now")
	}

	if stream.Status == Starting {
		return nil, errors.New("stream is about to start now")
	}

	if stream.Status == Started {
		return nil, errors.New("stream is already started")
	}

	//stream.Status = status
	return stream, nil
}

func (m *Manager) startStreaming(id int64, from int) error {
	whoSent := ""
	if from == FromClient {
		whoSent = "Rest API"
	} else if from == FromWatcher {
		whoSent = "Watcher"
	} else {
		whoSent = "unknown"
	}
	log.WithFields(log.Fields{
		"from": whoSent,
	}).Infof("received stream-%d start request", id)

	stream, err := m.canMakeStreamStatus(id, Started)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorf("cannot change stream-%d status to 'started", id)
		return err
	}

	if err := m.createStreamDir(stream); err != nil {
		return err
	}

	if err := m.cleanStreamDir(stream); err != nil {
		log.Warn("failed to clear streaming directories:", err)
	}

	go func() {
		if err := stream.start(); err != nil {
			log.WithFields(log.Fields{
				"id": id,
			}).Errorf("[manager] failed to start stream-%d: %s", id, err)
			return
		}
		log.WithFields(log.Fields{
			"id":  id,
			"url": stream.Uri,
		}).Errorf("stream-%d has been started", id)
	}()

	return nil
}

func (m *Manager) stopStreaming(id int64) error {
	stream := m.getStreamById(id)
	if stream == nil {
		return ErrorStreamNotFound
	}
	stream.stop()
	return nil
}

func (m *Manager) Stop() error {
	m.cancel()
	log.Debug("[manager] stopping all streams")
	for _, stream := range m.streams {
		//if stream.IsActive() {
		stream.stop()
		//}
	}

	return nil
}
func (m *Manager) startStreamWatcher() {
	go func() {
		log.WithFields(log.Fields{
			"interval(sec)": m.watcherCheckInterval.Seconds(),
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

func (m *Manager) checkStreams() {
	for id, stream := range m.streams {
		if !stream.Enabled {
			continue
		}
		if stream.IsActive() {
			continue
		}

		log.WithFields(log.Fields{}).Debugf("[watcher] since stream-%d is not running, restart it", id)
		if err := m.startStreaming(id, FromWatcher); err != nil {
			log.Error(err)
		}
	}

}
