package server

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/devplayg/rtsp-stream/common"
	"github.com/devplayg/rtsp-stream/streaming"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (m *Manager) deleteOldData(dataRetentionDays int) error {
	if dataRetentionDays < 0 {
		return nil
	}

	t := time.Now().In(common.Loc)
	targetTime := t.Add(time.Duration(dataRetentionDays*-24) * time.Hour)
	targetDate := targetTime.Format(common.DateFormat)
	log.WithFields(log.Fields{
		"now":               t.Format(time.RFC3339),
		"targetDate":        targetDate,
		"dataRetentionDays": dataRetentionDays,
	}).Debug("[manager] target date to delete")

	for _, s := range m.streams {
		err := m.deleteOldDataOfStream(s, targetDate)
		if err != nil {
			log.Error(err)
		}
	}

	return nil
}

func (m *Manager) deleteOldDataOfStream(s *streaming.Stream, date string) error {
	bucketName := []byte(fmt.Sprintf("%s%d", common.VideoBucketPrefix, s.Id))
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return nil
		}
		return b.ForEach(func(k, v []byte) error {
			result := strings.Compare(date, string(k))
			if result != 1 {
				return nil
			}
			dir := filepath.Join(m.server.config.Storage.RecordDir, strconv.FormatInt(s.Id, 10), string(k))
			log.WithFields(log.Fields{
				"streamId": s.Id,
				"key":      string(k),
				"dir":      dir,
			}).Debug("data is expired")
			var err error

			err = os.RemoveAll(dir)
			if err != nil {
				log.Error(err)
			}

			err = b.Delete(k)
			if err != nil {
				log.Error(err)
			}

			return nil
		})
	})
}

//func DeleteVideoRecordsBefore(dir, date string) error {
//	dir = filepath.ToSlash(dir)
//	log.WithFields(log.Fields{
//		"dir":  dir,
//		"date": date,
//	}).Debug("#### delete")
//
//	files, err := ioutil.ReadDir(dir)
//	if err != nil {
//		return err
//	}
//
//	for _, f := range files {
//		if !f.IsDir() {
//			continue
//		}
//		log.Debug(f.Name())
//	}
//
//	return nil
//}
