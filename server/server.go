package server

import (
	"encoding/json"
	"github.com/devplayg/hippo"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	engine     *hippo.Engine
	streamMap  sync.Map
	controller *Controller
}

func NewServer() *Server {
	server := Server{}
	server.controller = NewController(&server)
	return &server
}

func (s *Server) Start() error {
	logrus.Debug("RTSP Server is running")
	srv := &http.Server{
		Handler: s.controller.router,
		Addr:    "0.0.0.0:9000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go log.Fatal(srv.ListenAndServe())

	return nil
}

func (s *Server) Stop() error {
	logrus.Debug("RTSP server is stopping")

	// Save

	// Stop all streams
	//println("classifier is stopping")
	//err := c.DB.Close()
	//if err != nil {
	//    return err
	//}
	return nil
}

func (s *Server) getAllStreams() ([]byte, error) {
	streams := make([]*Stream, 0)
	s.streamMap.Range(func(k, v interface{}) bool {
		streams = append(streams, v.(*Stream))
		//fmt.Printf("key: %s, value: %s\n", k, v) // key: hoge, value: fuga
		return true
	})

	return json.Marshal(streams)
}
func (s *Server) AddStream(stream *Stream) error {
	_, ok := s.streamMap.Load(stream.Url)
	if ok {
		return errors.New("duplicate stream")
	}

	// stream := val.(Stream)
	s.streamMap.Store(stream.Url, stream)

	return nil
}
