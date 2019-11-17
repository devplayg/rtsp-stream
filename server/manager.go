package server

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/boltdb/bolt"
	"github.com/devplayg/hippo"
	"github.com/minio/highwayhash"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var StreamsKey = []byte("streams")

type Manager struct {
	server  *Server
	streams sync.Map
}

func NewManager(server *Server) *Manager {
	return &Manager{
		server: server,
	}
}

func (m *Manager) saveStreams() error {
	b, err := json.Marshal(m.getStreams())
	if err != nil {
		return err
	}

	return m.server.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(StreamBucket)
		return bucket.Put(StreamsKey, b)
	})
}

func (m *Manager) loadStreams() error {
	data, err := m.server.GetDbValue(StreamBucket, StreamsKey)
	if err != nil {
		return err
	}

	var streams []Stream
	err = json.Unmarshal(data, &streams)
	if err != nil {
		return err
	}

	m.streams = sync.Map{}
	for i := range streams {
		key := m.getStreamKey(&streams[i])
		m.streams.Store(key, &streams[i])
	}

	return nil
}

func (m *Manager) getStreams() []*Stream {
	streams := make([]*Stream, 0)
	m.streams.Range(func(k interface{}, v interface{}) bool {
		s := v.(*Stream)
		streams = append(streams, s)
		return true
	})

	return streams
}

func (m *Manager) getStreamById(id string) *Stream {
	val, ok := m.streams.Load(id)
	if !ok {
		return nil
	}

	return val.(*Stream)
}

func (m *Manager) getStreamKey(stream *Stream) string {
	hash := highwayhash.Sum128([]byte(stream.Uri), HashKey)
	return hex.EncodeToString(hash[:])
}

func (m *Manager) addStream(stream *Stream) error {
	key := m.getStreamKey(stream)
	_, ok := m.streams.Load(key)
	if ok {
		return ErrorDuplicatedStream
	}
	stream.Id = key
	stream.CmdType = NormalStream
	stream.LiveDir = filepath.ToSlash(filepath.Join(m.server.liveDir, stream.Id))
	stream.RecDir = filepath.ToSlash(filepath.Join(m.server.recDir, stream.Id))
	m.streams.Store(stream.Id, stream)

	return m.saveStreams()
}

func (m *Manager) deleteStream(id string) error {
	stream := m.getStreamById(id)
	if stream == nil {
		return ErrorStreamNotFound
	}

	m.streams.Delete(id)
	return m.saveStreams()
}

func (m *Manager) startStream(id string) error {
	stream := m.getStreamById(id)
	if stream == nil {
		return ErrorStreamNotFound
	}

	if stream.cmd != nil {
		proc, _ := os.FindProcess(stream.cmd.Process.Pid)
		if proc != nil {
			return errors.New("streaming is already working")
		}
	}

	//dir := filepath.Join(m.server.liveDir, stream.Id)
	err := hippo.EnsureDir(stream.LiveDir)
	if err != nil {
		return err
	}
	stream.cmd = GenerateStreamCommand(stream)
	go func() {
		err := stream.cmd.Run()
		if err != nil {
			log.Error(err)
			return
		}
	}()
	log.WithFields(log.Fields{
		"id":  stream.Id,
		"uri": stream.Uri,
	}).Info("streaming has been started")
	return nil
}

func (m *Manager) stopStream(id string) error {
	stream := m.getStreamById(id)
	if stream == nil {
		return ErrorStreamNotFound
	}

	if stream.cmd == nil {
		return nil
	}
	err := stream.cmd.Process.Kill()
	if err != nil {
		log.Debug("process message check: ", err)
		if strings.Contains(err.Error(), "process already finished") {
			return nil
		}
		if strings.Contains(err.Error(), "signal: killed") {
			return nil
		}

	}

	return nil
}
