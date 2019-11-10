package server

import (
	"github.com/devplayg/hippo"
	"github.com/sirupsen/logrus"
)

type Server struct {
	engine *hippo.Engine
	// router
	//
}

func (s *Server) Start() error {
	logrus.Debug("RTSP Server is running")

	url := "rtsp://admin:unisem1234@58.72.99.132:30101/Streaming/Channels/101/"
	stream := NewStreame(url)
	stream.Start()
	return nil
}

func (s *Server) Stop() error {
	logrus.Debug("RTSP server is stopping")
	//println("classifier is stopping")
	//err := c.DB.Close()
	//if err != nil {
	//    return err
	//}
	return nil
}
