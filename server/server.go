package server

import (
	"github.com/devplayg/hippo"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"time"
)

type Server struct {
	engine     *hippo.Engine
	controller *Controller
	manager    *Manager
}

func NewServer() *Server {
	server := &Server{}
	server.controller = NewController(server)
	server.manager = NewManager(server)
	return server
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
