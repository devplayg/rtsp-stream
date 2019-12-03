package streaming

import (
	"context"
	log "github.com/sirupsen/logrus"
	"time"
)

type Assistant struct {
	mpu8CaptureInterval time.Duration
	healthCheckInterval time.Duration
	ctx                 context.Context
	stream              *Stream
}

func NewAssistant(stream *Stream, ctx context.Context) *Assistant {
	return &Assistant{
		mpu8CaptureInterval: 1500 * time.Millisecond,
		healthCheckInterval: 4 * time.Second,
		stream:              stream,
		ctx:                 ctx,
	}
}

func (s *Assistant) init() error {
	return nil
}

func (s *Assistant) start() error {
	if err := s.init(); err != nil {
		return nil
	}

	go s.startCheckingStreamStatus()
	//go s.startMergingVideoFiles()
	log.WithFields(log.Fields{
		"stream_id": s.stream.Id,
	}).Debug("assistant has been started")

	return nil
}

func (s *Assistant) startCheckingStreamStatus() error {
	log.Debug("heatth")
	for {
		// just in case
		if s.stream.Status == Started && !s.stream.IsActive() {
			log.WithFields(log.Fields{}).Errorf("[stream-%d] status was 'started' but it wasn't alive.", s.stream.Id)
			s.stream.stop()
		}

		select {
		case <-time.After(s.healthCheckInterval):
		case <-s.ctx.Done():
			log.WithFields(log.Fields{
				"stream_id": s.stream.Id,
			}).Debug("health check of assistant has been stopped")
			return nil
		}
	}
}
