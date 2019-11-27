package streaming

import (
	"context"
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
		mergeInterval:        10 * time.Second,
		healthCheckInterval:  4 * time.Second,
		stream:               stream,
		ctx:                  ctx,
		date:                 time.Now().In(Loc).Format(DateFormat),
		lastSentMediaFileSeq: -1,
	}
}

func (s *Assistant) start() error {
	go s.startCheckingStreamStatus()
	go s.startMergingVideoFiles()
	log.WithFields(log.Fields{
		"stream_id": s.stream.Id,
	}).Debug("assistant has been started")

	return nil
}

func (s *Assistant) startCheckingStreamStatus() error {
	for {
		if s.stream.IsActive() != s.stream.Active {
			s.stream.Active = s.stream.IsActive()
			log.WithFields(log.Fields{
				"stream_id": s.stream.Id,
				"active":    s.stream.Active,
			}).Debug("stream status changed")

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
