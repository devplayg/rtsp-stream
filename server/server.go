package server

import (
	"github.com/boltdb/bolt"
	"github.com/devplayg/hippo"
	"github.com/devplayg/rtsp-stream/common"
	"github.com/minio/minio-go"
	log "github.com/sirupsen/logrus"
	"net/http"
	"path/filepath"
	"time"
)

var db *bolt.DB

type Server struct {
	engine     *hippo.Engine  // Server framework
	controller *Controller    // Controller
	manager    *Manager       // Stream manager
	addr       string         // Service address
	dbDir      string         // Database directory
	config     *common.Config // config
}

func NewServer(config *common.Config) *Server {
	server := &Server{
		dbDir:  "db",
		config: config,
		addr:   config.BindAddress,
	}

	return server
}

func (s *Server) Start() error {
	if err := s.init(); err != nil {
		return err
	}

	srv := &http.Server{
		Handler:      s.controller.router,
		Addr:         s.addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	absLiveDir, _ := filepath.Abs(s.config.Storage.LiveDir)
	absRecordDir, _ := filepath.Abs(s.config.Storage.RecordDir)
	log.WithFields(log.Fields{
		"liveDir":           absLiveDir,
		"recordDir":         absRecordDir,
		"bucket":            s.config.Storage.Bucket,
		"staticDir":         s.config.StaticDir,
		"dataRetentionDays": s.config.DataRetentionDays,
	}).Infof("[server] listening on %s", s.addr)
	go func() {
		log.Fatal(srv.ListenAndServe())
	}()

	return nil
}

func (s *Server) Stop() error {
	if err := s.manager.Stop(); err != nil {
		log.Error(err)
	}

	if err := db.Close(); err != nil {
		log.Error(err)
	}

	return nil
}

func (s *Server) SetEngine(e *hippo.Engine) {
	s.engine = e
}

func (s *Server) init() error {
	if err := s.initTimezone(); err != nil {
		return err
	}

	if err := s.initDatabase(); err != nil {
		return err
	}

	if err := s.initDirectories(); err != nil {
		return err
	}

	if err := s.initStorage(); err != nil {
		return err
	}

	if err := s.initManagerAndController(); err != nil {
		return err
	}

	return nil
}

func (s *Server) initManagerAndController() error {
	s.manager = NewManager(s)
	if err := s.manager.start(); err != nil {
		return err
	}

	s.controller = NewController(s)
	return nil
}

func (s *Server) initDirectories() error {
	if err := hippo.EnsureDir(s.config.Storage.LiveDir); err != nil {
		return err
	}

	if !s.config.Storage.Remote {
		if err := hippo.EnsureDir(s.config.Storage.RecordDir); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) initTimezone() error {
	if len(s.config.Timezone) < 1 {
		common.Loc = time.Local
		return nil
	}

	loc, err := time.LoadLocation(s.config.Timezone)
	if err != nil {
		return err
	}
	common.Loc = loc
	log.WithFields(log.Fields{
		"timezone": loc.String(),
	}).Debug("timezone has been set")
	return nil
}

func (s *Server) initDatabase() error {
	if err := hippo.EnsureDir(s.dbDir); err != nil {
		return err
	}
	var err error

	db, err = bolt.Open(filepath.Join(s.dbDir, "stream.db"), 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	defaultBuckets := [][]byte{common.StreamBucket, common.ConfigBucket}
	err = db.Update(func(tx *bolt.Tx) error {
		for _, b := range defaultBuckets {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"dir": s.dbDir,
	}).Debug("[server] database has been loaded")
	return nil
}

func (s *Server) initStorage() error {
	client, err := minio.New(s.config.Storage.Address, s.config.Storage.AccessKey, s.config.Storage.SecretKey, s.config.Storage.UseSSL)
	if err != nil {
		log.WithFields(log.Fields{
			"address":   s.config.Storage.Address,
			"accessKey": s.config.Storage.AccessKey,
		}).Error("failed to connect to object storage")
		return err
	}
	common.MinioClient = client

	if len(s.config.Storage.Bucket) > 0 {
		common.VideoRecordBucket = s.config.Storage.Bucket
	}
	return nil
}
