package streaming

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Assistant struct {
	mergeInterval       time.Duration // 1 min
	healthCheckInterval time.Duration
	ctx                 context.Context
	stream              *Stream
}

func NewAssistant(stream *Stream, ctx context.Context) *Assistant {
	return &Assistant{
		mergeInterval:       30 * time.Second,
		healthCheckInterval: 4 * time.Second,
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

func (s *Assistant) archiveLiveVideos() error {

	// Check if stream is active
	// Read *.ts files older than 20 seconds of last modification
	liveVideoFiles, err := GetRecentFilesInDir(s.stream.LiveDir, 20*time.Second)
	if err != nil {
		return err
	}
	if len(liveVideoFiles) < 1 {
		return nil
	}
	sort.SliceStable(liveVideoFiles, func(i, j int) bool {
		return liveVideoFiles[i].File.ModTime().Before(liveVideoFiles[j].File.ModTime())
	})

	var text string
	for _, f := range liveVideoFiles {
		path := filepath.Join(f.Dir, f.File.Name())
		text += fmt.Sprintf("file '%s'\n", path)
	}
	tempFile, err := ioutil.TempFile("", "stream")
	if err != nil {
		return err
	}
	_, err = tempFile.WriteString(text)
	tempFile.Close()

	outputFileName := liveVideoFiles[0].File.ModTime().Format("20060102_150405") + ".mp4"
	outputFilePath := filepath.Join(s.stream.RecDir, outputFileName)
	err = ArchiveLiveVideos(tempFile.Name(), outputFilePath)
	log.WithFields(log.Fields{
		"stream_id": s.stream.Id,
		"liveDir":   s.stream.LiveDir,
		"recDir":    s.stream.RecDir,
		"list":      tempFile.Name(),
		"output":    outputFilePath,
	}).Debug("assistant merged video files")
	for _, f := range liveVideoFiles {
		log.Debugf("    - %s", f.File.Name())
	}
	if err != nil {
		return err
	}

	s.removeLiveVideos(liveVideoFiles)
	return nil
}

func (s *Assistant) generateVideoList([]*LiveVideoFile) (string, error) {
	return "", nil
}

func (s *Assistant) removeLiveVideos(list []*LiveVideoFile) {
	for _, f := range list {
		path := filepath.Join(f.Dir, f.File.Name())
		os.Rename(path, filepath.Join(s.stream.RecDir, f.File.Name()))
		//if err := os.Remove(path); err != nil {
		//	log.Error(err)
		//}
	}
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
		if s.stream.Active { // wondory
			if err := s.archiveLiveVideos(); err != nil {
				log.Error(err)
			}
		}

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
