package server

import (
	"github.com/devplayg/rtsp-stream/common"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"time"
)

func (m *Manager) deleteOldDataBefore() error {
	t := time.Now()
	hours := -7
	dur := time.Duration(hours*-1) * time.Hour
	date := t.In(common.Loc).Add(dur)

	log.WithFields(log.Fields{
		"now":  t.In(common.Loc).Format(time.RFC3339),
		"date": date.Format(time.RFC3339),
	}).Debug("###################")

	for _, s := range m.streams {
		dir := filepath.Join(m.server.config.Storage.RecordDir, strconv.FormatInt(s.Id, 10))

		DeleteVideoRecordsBefore(dir, date.Format("20060102"))
	}

	// Delete old files in the storage directory

	// Delete old keys in the database

	// Delete old keys in the live database

	// Delete old keys in the live directory
	return nil
}

func DeleteVideoRecordsBefore(dir, date string) error {
	dir = filepath.ToSlash(dir)
	log.WithFields(log.Fields{
		"dir":  dir,
		"date": date,
	}).Debug("#### delete")

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		log.Debug(f.Name())
	}

	return nil
}
