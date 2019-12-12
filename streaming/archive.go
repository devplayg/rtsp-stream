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
	"strconv"
	"time"
)

func (m *Manager) startScheduler() error {
	m.scheduler = cron.New()

	//common.DB.Update(func(tx *bolt.Tx) error {
	//   bucket, err := tx.CreateBucketIfNotExists([]byte("video-1"))
	//   if err != nil {
	//       return err
	//   }
	//   //bucket.Put([]byte("20191201"), []byte{})
	//   //bucket.Put([]byte("20191202"), []byte{})
	//   bucket.Put([]byte("20191204"), []byte{})
	//   //bucket.Put([]byte("20191211"), []byte{})
	//   return nil
	//})
	//
	//common.DB.Update(func(tx *bolt.Tx) error {
	//   bucket, err := tx.CreateBucketIfNotExists([]byte("video-2"))
	//   if err != nil {
	//       return err
	//   }
	//   bucket.Put([]byte("20191204"), []byte{})
	//   return nil
	//})
	//
	//
	//common.DB.Update(func(tx *bolt.Tx) error {
	//   bucket, err := tx.CreateBucketIfNotExists([]byte("video-3"))
	//   if err != nil {
	//       return err
	//   }
	//   bucket.Put([]byte("20191203"), []byte{})
	//   bucket.Put([]byte("20191204"), []byte{})
	//   bucket.Put([]byte("20191205"), []byte{})
	//   return nil
	//})

	if err := m.startDelete([]int64{1}, time.Now().In(common.Loc).Add(24*time.Hour)); err != nil {
		log.Error(err)
	}

	_, err := m.scheduler.AddFunc("10 0 * * *", func() {
		t := time.Now().In(common.Loc)
		listToArchive, listToDelete := m.getStreamIdListToArchive()

		yesterday := t.Format(common.DateFormat)
		if err := m.startArchive(listToArchive, yesterday); err != nil {
			log.Error(err)
		}

		if err := m.startDelete(listToDelete, t); err != nil {
			log.Error(err)
		}
	})
	return err
}

func (m *Manager) getStreamIdListToArchive() ([]int64, []int64) {
	var listToArchive []int64
	var listToDelete []int64
	m.RLock()
	defer m.Unlock()
	for id, stream := range m.streams {
		if stream.Recording {
			listToArchive = append(listToArchive, id)
			continue
		}
		listToDelete = append(listToDelete, id)
	}

	return listToArchive, listToDelete
}

func (m *Manager) startArchive(streamIdList []int64, date string) error {
	if len(streamIdList) < 1 {
		return nil
	}
	for _, streamId := range streamIdList {
		liveDir := filepath.Join(m.server.config.Storage.LiveDir, strconv.FormatInt(streamId, 10))
		if err := m.archive(streamId, liveDir, date); err != nil {
			log.Error(err)
			continue
		}
		if err := m.writeArchiveHistory(streamId, date); err != nil {
			log.Error(err)
			continue
		}
	}

	return nil
}

func (m *Manager) startDelete(streamIdList []int64, t time.Time) error {
	if len(streamIdList) < 1 {
		return nil
	}
	for _, streamId := range streamIdList {
		liveDir := filepath.Join(m.server.config.Storage.LiveDir, strconv.FormatInt(streamId, 10))
		filesToDelete, err := common.ReadVideoFilesInDirNotOnDate(liveDir, t.Format(common.DateFormat), common.VideoFileExt)
		if err != nil {
			log.Error(err)
			continue
		}
		common.RemoveLiveFiles(liveDir, filesToDelete)
	}
	return nil
}

func (m *Manager) archive(streamId int64, liveDir string, date string) error {
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

	recordDir := filepath.ToSlash(filepath.Join(m.server.config.Storage.RecordDir, m.server.config.Storage.Bucket, strconv.FormatInt(streamId, 10), date))
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
	common.RemoveLiveFiles(liveDir, liveFiles)
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

func (m *Manager) writeArchiveHistory(streamId int64, date string) error {
	bucketName := []byte(common.VideoBucketPrefix + strconv.FormatInt(streamId, 10))
	return common.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(date), []byte{})
	})
}
