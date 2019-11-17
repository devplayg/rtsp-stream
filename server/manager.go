package server

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/devplayg/hippo"
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

func (m *Manager) save() error {
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
		m.streams.Store(streams[i].Id, &streams[i])
	}

	return nil
}

func (m *Manager) getStreams() []*Stream {
	streams := make([]*Stream, 0)
	m.streams.Range(func(k interface{}, v interface{}) bool {
		s := v.(*Stream)
		s.Active = s.IsActive()
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

func (m *Manager) setStream(stream *Stream) {
	stream.Id = GetHashString(stream.Uri)
	stream.CmdType = NormalStream
	stream.LiveDir = filepath.ToSlash(filepath.Join(m.server.liveDir, stream.Id))
	stream.RecDir = filepath.ToSlash(filepath.Join(m.server.recDir, stream.Id))
	stream.cmd = GenerateStreamCommand(stream)
}

func (m *Manager) addStream(stream *Stream) error {
	// Set the stream
	m.setStream(stream)

	// Check if the stream is duplicated
	if _, ok := m.streams.Load(stream.Id); ok {
		return ErrorDuplicatedStream
	}

	// Update stream map
	m.streams.Store(stream.Id, stream)

	if err := m.save(); err != nil {
		log.Error("failed to save stream")
		return err
	}

	if err := m.startStream(stream); err != nil {
		log.Error("failed to start stream")
		return err
	}

	return nil
}

func (m *Manager) deleteStream(id string) error {
	stream := m.getStreamById(id)
	if stream == nil {
		return ErrorStreamNotFound
	}

	err := m.removeStreamDir(stream)
	if err != nil {
		return err
	}

	m.streams.Delete(id)
	return m.save()
}

func (m *Manager) cleanStreamDir(stream *Stream) error {
	err := os.RemoveAll(stream.LiveDir)
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) removeStreamDir(stream *Stream) error {
	err := os.RemoveAll(stream.LiveDir)
	if err != nil {
		return err
	}
	err = os.RemoveAll(stream.RecDir)
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) createStreamDir(stream *Stream) error {
	if err := hippo.EnsureDir(stream.LiveDir); err != nil {
		return err
	}

	if err := hippo.EnsureDir(stream.RecDir); err != nil {
		return err
	}

	return nil
}

func (m *Manager) startStream(stream *Stream) error {
	m.setStream(stream)

	if err := m.cleanStreamDir(stream); err != nil {
		return err
	}

	if err := m.createStreamDir(stream); err != nil {
		return err
	}

	// Start process
	go func() {
		if err := stream.cmd.Run(); err != nil {
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

func (m *Manager) stopStreamProcess(id string) error {
	stream := m.getStreamById(id)
	if stream == nil {
		return ErrorStreamNotFound
	}

	if stream.cmd == nil {
		return nil
	}

	//err := stream.cmd.Process.Kill()
	err := stream.cmd.Process.Signal(os.Kill)
	log.Debug("check err: ", err)
	if err != nil {
		log.Debug("process message check: ", err)
		if strings.Contains(err.Error(), "process already finished") {
			return nil
		}
		if strings.Contains(err.Error(), "signal: killed") {
			return nil
		}
		if strings.Contains(err.Error(), "exit status 1") {
			return nil
		}
	}

	return err
}
