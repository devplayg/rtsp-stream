package server

import (
	"github.com/boltdb/bolt"
	"github.com/devplayg/hippo"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type Server struct {
	engine     *hippo.Engine
	controller *Controller
	manager    *Manager
	addr       string
	liveDir    string
	recDir     string
	db         *bolt.DB
}

func NewServer() *Server {
	server := &Server{
		addr:    "0.0.0.0:9000",
		liveDir: "f:/data/live/",
		recDir:  "f:/data/rec/",
	}

	//net.ResolveTCPAddr("tcp", tcpAddr)
	manager := NewManager(server)

	controller := NewController(server, manager)

	server.manager = manager
	server.controller = controller

	return server
}

func (s *Server) Start() error {
	err := s.init()
	if err != nil {
		return err
	}

	srv := &http.Server{
		Handler: s.controller.router,
		Addr:    s.addr,

		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.WithFields(log.Fields{
		"address": s.addr,
	}).Info("listen")
	go func() {
		log.Fatal(srv.ListenAndServe())
	}()

	return nil
}

func (s *Server) Stop() error {
	var err error
	err = s.db.Close()
	if err != nil {
		log.Error(err)
	}

	return nil
}

func (s *Server) SetEngine(e *hippo.Engine) {
	s.engine = e
}

func (s *Server) init() error {
	var err error

	err = s.initDatabase()
	if err != nil {
		return nil
	}
	log.Debug("database has been loaded")

	s.manager.loadStreams()

	return nil
}

func (s *Server) initDatabase() error {
	db, err := bolt.Open(s.engine.Config.Name+".db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	defaultBuckets := [][]byte{
		StreamBucket,
		ConfigBucket,
	}

	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, b := range defaultBuckets {
		_, err := tx.CreateBucketIfNotExists(b)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	s.db = db
	return nil
}

func (s *Server) GetDbValue(bucket, key []byte) ([]byte, error) {
	var data []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucket)
		data = bucket.Get(key)
		return nil
	})
	return data, err
}
