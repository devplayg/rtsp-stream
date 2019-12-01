package streaming

import (
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
	server  *Server
	streams map[int64]*Stream // Stream pool
	sync.RWMutex
}

func NewManager(server *Server) *Manager {
	return &Manager{
		server:  server,
		streams: make(map[int64]*Stream), /* key: id(int64), value: &stream */
	}
}

func (m *Manager) start() error {
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
			m.streams[stream.Id] = &stream
		}

		return nil
	})
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
	m.RLock()
	defer m.RUnlock()
	return m.streams[id]
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

	//err := os.RemoveAll(stream.liveDir)
	//if err != nil {
	//	return err
	//}
	return nil
}

func (m *Manager) removeStreamDir(stream *Stream) error {
	err := os.RemoveAll(stream.liveDir)
	if err != nil {
		return err
	}
	//err = os.RemoveAll(stream.RecDir)
	//if err != nil {
	//	return err
	//}
	return nil
}

func (m *Manager) createStreamDir(stream *Stream) error {
	stream.liveDir = filepath.ToSlash(filepath.Join(m.server.liveDir, strconv.FormatInt(stream.Id, 10)))

	if err := hippo.EnsureDir(stream.liveDir); err != nil {
		return err
	}

	return nil
}

//
//func (m *Manager) IsExistUri(uri string) bool {
//	hash := GetHashString(uri)
//	m.RLock()
//	for _, stream := range m.streams {
//		if stream.Hash == hash {
//			return true
//		}
//	}
//	m.RUnlock()
//	return false
//}

func (m *Manager) makeStreamStatus(id int64, status int) (*Stream, error) {
	m.Lock()
	defer m.Unlock()
	stream := m.streams[id]
	if stream == nil {
		return nil, ErrorStreamNotFound
	}
	if status == Stopped {
		if stream.Status == Stopped {
			return nil, errors.New("stream is already stopped")
		}
		if stream.Status == Stopping {
			return nil, errors.New("stream is stopping now")
		}

		if stream.Status == Starting {
			return nil, errors.New("stream is about to start now")
		}

		stream.Status = status
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

	stream.Status = status
	return stream, nil
}

func (m *Manager) startStreaming(id int64) error {
	log.WithFields(log.Fields{
		"id": id,
	}).Debug("received stream start request")

	stream, err := m.makeStreamStatus(id, Started)
	if err != nil {
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
			}).Error("failed to start stream: ", err)
			return
		}
		log.Debug("stream started")
	}()

	return nil
}

func (m *Manager) stopStreaming(id int64) error {
	stream := m.getStreamById(id)
	if stream == nil {
		return ErrorStreamNotFound
	}

	if err := stream.stop(); err != nil {
		return err
	}

	//log.WithFields(log.Fields{
	//	"stream_id": id,
	//}).Debugf("stream has been stopped")
	return nil
}

func (m *Manager) Stop() error {
	// Stop all running streamings
	for _, stream := range m.streams {
		err := stream.stop()
		log.WithFields(log.Fields{
			"stream_id": stream.Id,
			"uri":       stream.Uri,
			"err":       err,
		}).Debugf("stream stop")
	}
	return nil
}

//
//func (m *Manager) printStream(stream *Stream) {
//	log.Debug("===================================================")
//	log.Debugf("id: %d", stream.Id)
//	log.Debugf("hash: %d", stream.hash)
//	log.Debugf("uri: %s", stream.Uri)
//	log.Debugf("active: %s", stream.Active)
//	log.Debugf("recording: %s", stream.Recording)
//	log.Debug("===================================================")
//}

// NEED STREAM RECONNECTION
