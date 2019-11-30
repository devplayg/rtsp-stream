package streaming

import (
	"errors"
	"github.com/boltdb/bolt"
	"github.com/devplayg/hippo"
	"github.com/minio/minio-go"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var (
	DB          *bolt.DB
	Loc         *time.Location
	MinioClient *minio.Client
)

type Server struct {
	engine     *hippo.Engine
	controller *Controller
	manager    *Manager
	addr       string
	liveDir    string
	recDir     string
	config     *Config
}

func NewServer(config *Config) *Server {
	server := &Server{
		config:  config,
		addr:    config.BindAddress,
		liveDir: config.Storage.Live,
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
	if err := s.manager.Stop(); err != nil {
		log.Error(err)
	}

	if err := DB.Close(); err != nil {
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
	Loc = loc

	if len(s.config.Storage.Bucket) > 0 {
		VideoRecordBucket = s.config.Storage.Bucket
	}

	// Set controller
	s.controller = NewController(s)

	MinioClient, err = minio.New(s.config.Storage.Address, s.config.Storage.AccessKey, s.config.Storage.SecretKey, s.config.Storage.UseSSL)
	if err != nil {
		return errors.New("failed to connect to object storage; "+ err.Error())
	}

	return nil
}

func (s *Server) initDatabase() error {
	dbName := s.engine.Config.Name + ".db"
	db, err := bolt.Open(dbName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	defaultBuckets := [][]byte{StreamBucket, TransmissionBucket, ConfigBucket}
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

	DB = db
	log.WithFields(log.Fields{
		"db": dbName,
	}).Debug("BoltDB has been loaded")
	return nil
}

func (s *Server) GetDbValue(bucket, key []byte) ([]byte, error) {
	var data []byte
	err := DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucket)
		data = bucket.Get(key)
		return nil
	})
	return data, err
}

func (s *Server) initStor() error {