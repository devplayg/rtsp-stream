package server

import (
	log "github.com/sirupsen/logrus"
	"time"
)

type Assistant struct {
	interval time.Duration // 1 min
	stream   *Stream
}

func NewAssistant(stream *Stream) *Assistant {
	return &Assistant{
		interval: 5 * time.Second,
		stream:   stream,
	}
}

func (s *Assistant) start() error {

	//go s.run()
	//go func() {
	//    done <- s.run()
	//}()
	//
	//select {
	//case result := <-done:
	//    return result, nil
	//case <-ctx.Done():
	//    return "Fail", ctx.Err()
	//}

	log.WithFields(log.Fields{
		"stream_id": s.stream.Id,
	}).Debug("stream assistant has been started")

	return nil
}

func (s *Assistant) stop() error {
	log.WithFields(log.Fields{
		"stream_id": s.stream.Id,
	}).Debug("stream assistant has been stopped")

	return nil
}

func (s *Assistant) run() error {
	for {
		//        // Read *.ts files older than 20 seconds of last modification
		//        tsFiles, startTime, err := a.geLiveTsFiles(20 * time.Second)
		//
		//        // Merge *.ts files and save as timestamp.mp4(timestamp fo the first ts file)
		//        mergePath, err := a.mergeTsFiles(startTime)
		//        if err != nil {
		//            log.Error(err)
		//            time.Sleep(a.interval)
		//            continue
		//        }
		//
		//        // Remove *.ts files
		//        for _, path := range tsFiles {
		//            if err := os.Remove(path); err != nil {
		//                log.Error(err)
		//            }
		//        }
		//z

		log.WithFields(log.Fields{
			"stream_id": s.stream.Id,
		}).Debug("stream assistant is doing something..")
		time.Sleep(s.interval)
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
