package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/devplayg/hippo"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"math"
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

//
//func (s *Assistant) stop() error {
//	log.WithFields(log.Fields{
//		"stream_id": s.stream.Id,
//	}).Debug("stream assistant has been stopped")
//
//	return nil
//}

func (s *Assistant) getLiveVideoFiles() ([]*LiveVideoFile, error) {
	// Read *.ts files older than 20 seconds of last modification
	liveVideoFiles, err := GetRecentFilesInDir(s.stream.LiveDir, 20*time.Second)
	if err != nil {
		return nil, err
	}

	// Sort ts files by modification time (live9.ts, live10.ts, live11.ts ...)
	sort.SliceStable(liveVideoFiles, func(i, j int) bool {
		return liveVideoFiles[i].File.ModTime().Before(liveVideoFiles[j].File.ModTime())
	})

	return liveVideoFiles, nil
}

//func (s *Assistant) generateFileListForUseWithFfmpeg(liveVideoFiles []*LiveVideoFile) (*os.File, error) {
//    var text string
//    for _, f := range liveVideoFiles {
//        path := filepath.ToSlash(filepath.Join(f.Dir, f.File.Name()))
//        text += fmt.Sprintf("file %s\n", path)
//    }
//    tempFile, err := ioutil.TempFile("", "stream")
//    if err != nil {
//        return nil, err
//    }
//    defer tempFile.Close()
//    _, err = tempFile.WriteString(text)
//    if err != nil {
//        return nil, err
//    }
//
//    return tempFile, nil
//}

func (s *Assistant) mergeLiveVideoFiles(liveVideoFiles []*LiveVideoFile, file *os.File) (*VideoRecord, error) {

	// Merge live *.ts files to record *.ts files
	ext := ".ts"
	videoRecord := NewVideoRecord(liveVideoFiles[0].File.ModTime(), Loc, ext)
	date := liveVideoFiles[0].File.ModTime().In(Loc).Format("20060102")
	if err := hippo.EnsureDir(filepath.Join(s.stream.RecDir, date)); err != nil {
		return nil, err
	}
	outputFilePath := filepath.Join(s.stream.RecDir, date, videoRecord.Name)
	duration, err := MergeLiveVideoFiles(file.Name(), outputFilePath)
	log.WithFields(log.Fields{
		"stream_id":       s.stream.Id,
		"liveDir":         s.stream.LiveDir,
		"recDir":          s.stream.RecDir,
		"list":            file.Name(),
		"output":          outputFilePath,
		"archivingResult": err,
		"duration":        duration,
	}).Debug("assistant merged video files")
	for _, f := range liveVideoFiles {
		log.Debugf("    - %s", f.File.Name())
	}
	videoRecord.Duration = duration
	return videoRecord, err
}

func (s *Assistant) archiveLiveVideos() error {

	// Get live video files
	liveVideoFiles, err := s.getLiveVideoFiles()
	if err != nil || len(liveVideoFiles) < 1 {
		return err
	}

	// Generate file list of live video files for use with ffmpeg
	fileList, err := GenerateLiveVideoFileListForUseWithFfmpeg(liveVideoFiles)
	if err != nil {
		return err
	}
	defer os.Remove(fileList.Name())

	// Merge live video files
	videoFile, err := s.mergeLiveVideoFiles(liveVideoFiles, fileList)
	if err != nil {
		return err
	}
	//spew.Dump(videoFile)

	if err := s.saveVideoRecord(videoFile); err != nil {
		return err
	}

	// update m3u8
	if err := s.updateM3u8(videoFile); err != nil {
		return err
	}

	s.removeLiveVideos(liveVideoFiles)
	return nil
}

func (s *Assistant) updateM3u8(videoRecord *VideoRecord) error {
	bucketName := GetVideRecordBucket(videoRecord, s.stream.Id)
	//videoRecords := make([]*VideoRecord, 0)
	var maxTargetDuration float32
	m3u8Header := GetM3u8Header()
	var body string

	err := DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			var videoRecord VideoRecord
			err := json.Unmarshal(k, &videoRecord)
			if err != nil {
				log.Error(err)
				continue
			}
			//spew.Dump(videoRecord)

			//videoRecords = append(videoRecords, &videoRecord)
			if videoRecord.Duration > maxTargetDuration {
				maxTargetDuration = videoRecord.Duration
			}
			body += fmt.Sprintf("#EXTINF:%.6f,\n", videoRecord.Duration)
			body += videoRecord.Name + "\n"
		}

		return nil
	})
	if err != nil {
		return err
	}

	m3u8Header += fmt.Sprintf("#EXT-X-TARGETDURATION:%.0f\n", math.Ceil(float64(maxTargetDuration)))
	m3u8Footer := "#EXT-X-ENDLIST"
	m3u8 := m3u8Header + body + m3u8Footer

	date := time.Unix(videoRecord.UnixTime, 0).In(Loc).Format("20060102")
	outputFilePath := filepath.Join(s.stream.RecDir, date, "index.m3u8")
	return ioutil.WriteFile(outputFilePath, []byte(m3u8), 0644)
}

func (s *Assistant) saveVideoRecord(videoRecord *VideoRecord) error {
	bucketName := GetVideRecordBucket(videoRecord, s.stream.Id)

	return DB.Update(func(tx *bolt.Tx) error {

		// Create bucket
		if _, err := tx.CreateBucketIfNotExists(bucketName); err != nil {
			return err
		}

		key, err := json.Marshal(videoRecord)
		if err != nil {
			return err
		}

		bucket := tx.Bucket(bucketName)
		return bucket.Put(key, []byte{})
	})
}

//func (s *Assistant) generateVideoList([]*LiveVideoFile) (string, error) {
//    return "", nil
//}

func (s *Assistant) removeLiveVideos(list []*LiveVideoFile) {
	for _, f := range list {
		path := filepath.Join(f.Dir, f.File.Name())
		//os.Rename(path, filepath.Join(s.stream.RecDir, f.File.Name()))
		if err := os.Remove(path); err != nil {
			log.Error(err)
		}
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
