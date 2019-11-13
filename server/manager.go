package server

import (
	"encoding/json"
	"errors"
	"sync"
)

type Manager struct {
	server    *Server
	streamMap sync.Map
}

func NewManager(server *Server) *Manager {
	return &Manager{
		server: server,
	}
}

func (m *Manager) getAllStreams() ([]byte, error) {
	streams := make([]*Stream, 0)
	m.streamMap.Range(func(k, v interface{}) bool {
		streams = append(streams, v.(*Stream))
		//fmt.Printf("key: %s, value: %s\n", k, v) // key: hoge, value: fuga
		return true
	})

	return json.Marshal(streams)
}

func (m *Manager) AddStream(stream *Stream) error {
	_, ok := m.streamMap.Load(stream.URI)
	if ok {
		return errors.New("duplicate stream")
	}

	// stream := val.(Stream)
	m.streamMap.Store(stream.URI, stream)

	return nil
}
