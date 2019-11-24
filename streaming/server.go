package streaming

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
	config     *Config
	loc        *time.Location
}

func NewServer(config *Config) *Server {
	server := &Server{
		config:  config,
		addr:    config.BindAddress,
		liveDir: config.Storage.Live,
		recDir:  config.Storage.Recording,
	}

	return server
}

func (s *Server) Start() error {
	if err := s.init(); err != nil {
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
	if err := s.db.Close(); err != nil {
		log.Error(err)
	}

	return nil
}

func (s *Server) SetEngine(e *hippo.Engine) {
	s.engine = e
}

func (s *Server) init() error {

	// Initialize database
	if err := s.initDatabase(); err != nil {
		return nil
	}

	// Set manager
	s.manager = NewManager(s)
	if err := s.manager.load(); err != nil {
		return err
	}

	// Init timezone
	loc, err := time.LoadLocation(s.config.Timezone)
	if err != nil {
		return err
	}
	s.loc = loc

	// Set controller
	s.controller = NewController(s)

	return nil
}

func (s *Server) initDatabase() error {
	dbName := s.engine.Config.Name + ".db"
	db, err := bolt.Open(dbName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	defaultBuckets := [][]byte{StreamBucket, ConfigBucket}
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, b := range defaultBuckets {
		if _, err := tx.CreateBucketIfNotExists(b); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	s.db = db
	log.WithFields(log.Fields{
		"db": dbName,
	}).Debug("BoltDB has been loaded")
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
