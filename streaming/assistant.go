package streaming

import (
	"context"
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/devplayg/rtsp-stream/utils"
	log "github.com/sirupsen/logrus"
	"time"
)

var (
	VideoFilePrefix     = "media"
	LiveVideoFilePrefix = "live"
	VideoFileExt        = ".ts"
)

type Assistant struct {
	mergeInterval        time.Duration // 1 min
	healthCheckInterval  time.Duration
	ctx                  context.Context
	stream               *Stream
	date                 string
	lastSentMediaFileSeq int
	lastSentHash         []byte
	lastSentSize         int64
}

func NewAssistant(stream *Stream, ctx context.Context) *Assistant {
	return &Assistant{
		mergeInterval:        60 * time.Second,
		healthCheckInterval:  4 * time.Second,
		stream:               stream,
		ctx:                  ctx,
		date:                 time.Now().In(Loc).Format(DateFormat),
		lastSentMediaFileSeq: -1,
	}
}

func (s *Assistant) init() error {
	err := DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(TransmissionBucket)
		data := b.Get(utils.Int64ToBytes(s.stream.Id))
		if data == nil {
			return nil
		}

		var result TransmissionResult
		err := json.Unmarshal(data, &result)
		if err != nil {
			log.Error(err)
			return nil
		}

		if time.Now().In(Loc).Format(DateFormat) == result.Date {
			s.lastSentMediaFileSeq = result.Seq
			s.lastSentHash = result.Hash
		}
		log.WithFields(log.Fields{
			"stream_id": s.stream.Id,
			"seq":       result.Seq,
			"hash":      string(result.Hash),
			"size":      result.Size,
		}).Debugf("[stream-%d] detected last tx result", s.stream.Id)
		return nil
	})

	return err
}

func (s *Assistant) start() error {
	if err := s.init(); err != nil {
		return nil
	}

	go s.startCheckingStreamStatus()
	go s.startMergingVideoFiles()
	log.WithFields(log.Fields{
		"stream_id": s.stream.Id,
	}).Debug("assistant has been started")

	return nil
}

func (s *Assistant) startCheckingStreamStatus() error {
	for {
		if s.stream.Status == Started {
			if !s.stream.IsActive() {
				if err := s.stream.stop(); err != nil {
					log.WithFields(log.Fields{
						"id": s.stream.Id,
					}).Error("stream status was 'started' but it was not alive. failed to stop stream")
				}
			}
		}

		select {
		case <-time.After(s.healthCheckInterval):
		case <-s.ctx.Done():
			log.WithFields(log.Fields{
				"stream_id": s.stream.Id,
			}).Debug("assistant stopped checking stream status")
			return nil
		}
	}
}

func (s *Assistant) startMergingVideoFiles() error {
	for {
		//if s.stream.Active { // wondory
		if err := s.archiveLiveVideos(); err != nil {
			log.Error(err)
		}
		//}

		select {
		case <-time.After(s.mergeInterval):
		case <-s.ctx.Done():
			log.WithFields(log.Fields{
				"stream_id": s.stream.Id,
			}).Debug("assistant stopped merging video files")
			return nil
		}
	}
}
