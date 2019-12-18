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

func (m *Manager) canArchive() bool {
	m.RLock()
	defer m.RUnlock()
	if m.onArchiving {
		return false
	}
	m.onArchiving = true
	return true
}

func (m *Manager) testScheduler() error {
	targetDate := time.Now().In(common.Loc).Add(-24 * time.Hour).Format(common.DateFormat) // Yesterday
	if err := m.startToArchiveVideos(targetDate); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (m *Manager) startScheduler() error {
	scheduler := cron.New(cron.WithLocation(common.Loc))

	_, err := scheduler.AddFunc("10 0 * * *", func() {
		if !m.canArchive() {
			log.Debug("archiving is already running")
			return
		}
		defer func() {
			m.RLock()
			m.onArchiving = false
			m.RUnlock()
		}()

		targetDate := time.Now().In(common.Loc).Add(-24 * time.Hour).Format(common.DateFormat) // Yesterday
		if err := m.startToArchiveVideos(targetDate); err != nil {
			log.Error(err)
			return
		}

	})

	scheduler.Start()
	m.scheduler = scheduler

	return err
}

func (m *Manager) startToArchiveVideos(targetDate string) error {
	t := time.Now()
	streamIdListToArchive, streamIdListNotToArchive := m.getStreamIdListToArchive()
	log.WithFields(log.Fields{
		"targetDate":               targetDate,
		"streamsCountToArchive":    len(streamIdListToArchive),
		"streamsCountNotToArchive": len(streamIdListNotToArchive),
	}).Debug("[manager] archiving is about to start")

	if len(streamIdListToArchive) > 0 {
		if err := m.startToArchiveVideosOnDate(streamIdListToArchive, targetDate); err != nil {
			log.Error("failed to archive videos")
			return err
		}
	}
	if err := m.startToDeleteVideosNotToBeArchived(streamIdListNotToArchive, targetDate); err != nil {
		log.Error("failed to delete videos")
		return err
	}

	log.WithFields(log.Fields{
		"duration(sec)": time.Since(t).Seconds(),
		"targetDate":    targetDate,
	}).Debug("[manager] archiving has been finished")
	return m.server.PutDataInDB(common.ConfigBucket, common.LastArchivingDateKey, []byte(targetDate))
}

func (m *Manager) getStreamIdListToArchive() ([]int64, []int64) {
	var streamIdListToArchive []int64
	var streamIdListNotToArchive []int64
	m.RLock()
	defer m.RUnlock()
	for id, stream := range m.streams {
		if stream.Recording {
			streamIdListToArchive = append(streamIdListToArchive, id)
			continue
		}
		streamIdListNotToArchive = append(streamIdListNotToArchive, id)
	}

	return streamIdListToArchive, streamIdListNotToArchive
}

func (m *Manager) startToArchiveVideosOnDate(streamIdList []int64, date string) error {
	var result error
	for _, streamId := range streamIdList {
		liveDir := filepath.Join(m.server.config.Storage.LiveDir, strconv.FormatInt(streamId, 10))
		if err := m.archive(streamId, liveDir, date); err != nil {
			log.Error(err)
			result = err
			continue
		}
		if err := m.writeVideoArchivingHistory(streamId, date); err != nil {
			log.Error(err)
			result = err
			continue
		}

		//if err := m.deleteLiveDataOnStreamDB(streamId, date); err != nil {
		//	log.Error(err)
		//	result = err
		//	continue
		//}
	}

	return result
}

func (m *Manager) deleteLiveDataOnStreamDB(streamId int64, date string) error {
	stream := m.getStreamById(streamId)
	if stream == nil {
		return errors.New(fmt.Sprintf("invalid stream id: %d", streamId))
	}

	return stream.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(date))
		if b == nil {
			return errors.New(fmt.Sprintf("streamId: %d, invalid bucket: %s", streamId, date))
		}

		return b.DeleteBucket([]byte(date))
	})
}

func (m *Manager) archive(streamId int64, liveDir string, date string) error {
	liveFiles, err := common.ReadVideoFilesOnDateInDir(liveDir, date, common.VideoFileExt)
	if err != nil {
		return err
	}

	if len(liveFiles) < 1 {
		log.WithFields(log.Fields{
			"date":     date,
			"dir":      liveDir,
			"streamId": streamId,
		}).Debug("no video files to archive")
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
		"date": date,
		"dir":  liveDir,
	}).Debugf("[manager] found %d video files in stream-%d; merging video files..", streamId, len(liveFiles))
	err = common.MergeLiveVideoFiles(listFilePath, filepath.Join(recordDir, common.LiveM3u8FileName), m.server.config.HlsOptions.SegmentTime)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"date":     date,
		"dir":      liveDir,
		"streamId": streamId,
		"count":    len(liveFiles),
		"duration": time.Since(t).Seconds(),
	}).Debug("[manager] completed merging video files")

	// common.RemoveLiveFiles(liveDir, liveFiles) // wondory
	return err
}

func (m *Manager) startToDeleteVideosNotToBeArchived(streamIdList []int64, targetDate string) error {
	if len(streamIdList) < 1 {
		return nil
	}
	for _, streamId := range streamIdList {
		liveDir := filepath.Join(m.server.config.Storage.LiveDir, strconv.FormatInt(streamId, 10))
		filesToDelete, err := common.ReadVideoFilesOnDateInDir(liveDir, targetDate, common.VideoFileExt)
		if err != nil {
			log.WithFields(log.Fields{
				"streamId":   streamId,
				"targetDate": targetDate,
				"dir":        liveDir,
			})
			log.Error("failed to remove unnecessary video files")
			continue
		}
		deleted := common.RemoveLiveFiles(liveDir, filesToDelete)
		log.WithFields(log.Fields{
			"streamId":   streamId,
			"targetDate": targetDate,
			"dir":        liveDir,
			"deleted":    deleted,
		}).Debug("[manater] removed unnecessary video files")
	}
	return nil
}

func (m *Manager) writeLiveFileListToText(liveDir string, files []os.FileInfo, recordDir string) (string, error) {
	var text string
	for _, f := range files {
		path, _ := filepath.Abs(filepath.ToSlash(filepath.Join(liveDir, f.Name())))
		text += fmt.Sprintf("file '%s'\n", path)
	}

	if len(text) < 1 {
		return "", errors.New("no data")
	}

	f, err := ioutil.TempFile(recordDir, "list")
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = f.WriteString(text)
	return f.Name(), err
}

func (m *Manager) writeVideoArchivingHistory(streamId int64, date string) error {
	bucketName := []byte(common.VideoBucketPrefix + strconv.FormatInt(streamId, 10))
	return common.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(date), []byte{})
	})
}

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
