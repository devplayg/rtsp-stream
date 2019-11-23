package streaming

import (
	"context"
	log "github.com/sirupsen/logrus"
	"time"
)

type Assistant struct {
	interval            time.Duration // 1 min
	healthCheckInterval time.Duration
	ctx                 context.Context
	stream              *Stream
}

func NewAssistant(stream *Stream, ctx context.Context) *Assistant {
	return &Assistant{
		interval:            5 * time.Second,
		healthCheckInterval: 8 * time.Second,
		stream:              stream,
		ctx:                 ctx,
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

func (s *Assistant) stop() error {
	log.WithFields(log.Fields{
		"stream_id": s.stream.Id,
	}).Debug("stream assistant has been stopped")

	return nil
}

func (s *Assistant) mergeVideoFiles() error {
	log.WithFields(log.Fields{
		"stream_id": s.stream.Id,
		"liveDir":   s.stream.LiveDir,
		"recDir":    s.stream.RecDir,
	}).Debug("assistant merged video files")

	//for {
	//	select {
	//	case <-time.After(s.interval):
	//	case <-s.ctx.Done():
	//		log.WithFields(log.Fields{
	//			"stream_id": s.stream.Id,
	//		}).Debug("assistant has been stopped video merging")
	//		return nil
	//	}
	//}
	return nil
}

func (s *Assistant) startCheckingStreamStatus() error {
	for {
		s.stream.Active = s.stream.IsActive()
		log.WithFields(log.Fields{
			"stream_id": s.stream.Id,
			"active":    s.stream.Active,
		}).Debug("assistant checked stream status")

		select {
		case <-time.After(s.healthCheckInterval):
		case <-s.ctx.Done():
			log.WithFields(log.Fields{
				"stream_id": s.stream.Id,
			}).Debug("assistant stopped checking stream status")
			return nil
		}
	}

	return nil
}

func (s *Assistant) startMergingVideoFiles() error {
	for {
		if s.stream.Active {
			if err := s.mergeVideoFiles(); err != nil {
				log.Error(err)
			}
		}
		// Check if stream is active
		//// Read *.ts files older than 20 seconds of last modification
		//tsFiles, startTime, err := a.geLiveTsFiles(20 * time.Second)
		//
		//// Merge *.ts files and save as timestamp.mp4(timestamp fo the first ts file)
		//mergePath, err := a.mergeTsFiles(startTime)
		//if err != nil {
		//log.Error(err)
		//time.Sleep(a.interval)
		//continue
		//}
		//
		//// Remove *.ts files
		//for _, path := range tsFiles {
		//if err := os.Remove(path); err != nil {
		//log.Error(err)
		//}
		//}

		// time.Sleep(s.interval)
		select {
		case <-time.After(s.interval):
		case <-s.ctx.Done():
			log.WithFields(log.Fields{
				"stream_id": s.stream.Id,
			}).Debug("assistant stopped merging video files")
			return nil
		}
	}

	return nil
}

//
//func (a *Assistant) mergeTsFiles(list []string) ([]string {
//    // merge
//
//    // Take snapshot
//
//    return nil
//}
//
//func (a *Assistant) geLiveTsFiles(dur time.Duration) []string {
//
//    // Sort *.ts files
//    return nil
//}

// Boltdb

/*
	streams
		key: id
		val: stream information

	records
		key: id

*/
