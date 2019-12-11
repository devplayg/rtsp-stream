package streaming

import (
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/devplayg/hippo"
	"github.com/devplayg/rtsp-stream/common"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func (m *Manager) startScheduler() error {
	m.scheduler = cron.New()

	// common.DB.Update(func(tx *bolt.Tx) error {
	//    bucket, err := tx.CreateBucketIfNotExists([]byte("video-1"))
	//    if err != nil {
	//        return err
	//    }
	//    bucket.Put([]byte("20191201"), []byte{})
	//    bucket.Put([]byte("20191202"), []byte{})
	//    bucket.Put([]byte("20191203"), []byte{})
	//    bucket.Put([]byte("20191211"), []byte{})
	//    return nil
	//})
	//
	// common.DB.Update(func(tx *bolt.Tx) error {
	//    bucket, err := tx.CreateBucketIfNotExists([]byte("video-2"))
	//    if err != nil {
	//        return err
	//    }
	//    bucket.Put([]byte("20191206"), []byte{})
	//    return nil
	//})
	//
	//
	// common.DB.Update(func(tx *bolt.Tx) error {
	//    bucket, err := tx.CreateBucketIfNotExists([]byte("video-3"))
	//    if err != nil {
	//        return err
	//    }
	//    bucket.Put([]byte("20191203"), []byte{})
	//    bucket.Put([]byte("20191204"), []byte{})
	//    bucket.Put([]byte("20191205"), []byte{})
	//    return nil
	//})

	_, err := m.scheduler.AddFunc("10 0 * * *", func() {
		yesterday := time.Now().In(common.Loc).Format(common.DateFormat)

		if err := m.startArchive(yesterday); err != nil {
			log.Error(err)
		}
	})
	return err
}

func (m *Manager) startArchive(date string) error {
	dirs, err := ioutil.ReadDir(m.server.config.Storage.LiveDir)
	if err != nil {
		return err
	}

	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}

		liveDir := filepath.ToSlash(filepath.Join(m.server.config.Storage.LiveDir, d.Name())) // live/1/
		if err := m.archive(liveDir, date, d.Name()); err != nil {
			log.Error(err)
			continue
		}

		if err := m.writeArchiveHistory(d.Name(), date); err != nil {

		}
	}
	return nil
}

func (m *Manager) archive(liveDir, date, streamId string) error {
	liveFiles, err := common.ReadVideoFilesInDirOnDate(liveDir, date, common.VideoFileExt)
	if err != nil {
		return err
	}

	if len(liveFiles) < 1 {
		log.WithFields(log.Fields{
			"date":     date,
			"dir":      liveDir,
			"streamId": streamId,
		}).Debug("no video files")
		return nil
	}
	sort.SliceStable(liveFiles, func(i, j int) bool {
		return liveFiles[i].ModTime().Before(liveFiles[j].ModTime())
	})

	recordDir := filepath.ToSlash(filepath.Join(m.server.config.Storage.RecordDir, m.server.config.Storage.Bucket, streamId, date))
	if err := hippo.EnsureDir(recordDir); err != nil {
		return err
	}
	listFilePath, err := m.writeLiveFileListToText(liveDir, liveFiles, recordDir)
	if err != nil {
		return err
	}

	t := time.Now()
	log.WithFields(log.Fields{
		"date":     date,
		"dir":      liveDir,
		"streamId": streamId,
	}).Debugf("found %d video files; merging video files..", len(liveFiles))
	err = common.MergeLiveVideoFiles(listFilePath, filepath.Join(recordDir, common.LiveM3u8FileName))
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"date":     date,
		"dir":      liveDir,
		"streamId": streamId,
		"count":    len(liveFiles),
		"duration": time.Since(t).Seconds(),
	}).Debug("completed merging video files")
	//common.RemoveLiveFiles(liveDir, liveFiles)
	return err
}

func (m *Manager) writeLiveFileListToText(liveDir string, files []os.FileInfo, tempDir string) (string, error) {
	var text string
	for _, f := range files {
		path, _ := filepath.Abs(filepath.ToSlash(filepath.Join(liveDir, f.Name())))
		text += fmt.Sprintf("file '%s'\n", path)
	}

	if len(text) < 1 {
		return "", errors.New("no data")
	}

	f, err := ioutil.TempFile(tempDir, "list")
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = f.WriteString(text)
	return f.Name(), err
}

func (m *Manager) writeArchiveHistory(streamId, date string) error {
	bucketName := []byte(common.VideoBucketPrefix + streamId)
	return common.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(date), []byte{})
	})
}
