package store

import (
	"errors"
	"fmt"
	"github.com/devplayg/hippo"
	"github.com/devplayg/rtsp-stream/common"
	"github.com/minio/minio-go"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Alba struct {
	engine    *hippo.Engine
	config    *common.Config
	scheduler *cron.Cron
}

func NewAlba(config *common.Config) *Alba {
	server := &Alba{
		config: config,
	}

	return server
}

func (a *Alba) Start() error {
	if err := a.init(); err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"loc": common.Loc,
	}).Debug()

	date := "20191204"
	if err := a.startArchive(date); err != nil {
		return err
	}

	return nil
}

func (a *Alba) Stop() error {
	return nil
}

func (a *Alba) SetEngine(e *hippo.Engine) {
	a.engine = e
}

func (a *Alba) init() error {

	if err := a.initTimezone(); err != nil {
		return err
	}

	if err := a.initStorage(); err != nil {
		return err
	}

	return nil
}

func (a *Alba) initTimezone() error {
	if len(a.config.Timezone) < 1 {
		common.Loc = time.Local
		return nil
	}

	loc, err := time.LoadLocation(a.config.Timezone)
	if err != nil {
		return err
	}
	common.Loc = loc
	return nil
}

func (a *Alba) initStorage() error {
	client, err := minio.New(a.config.Storage.Address, a.config.Storage.AccessKey, a.config.Storage.SecretKey, a.config.Storage.UseSSL)
	if err != nil {
		log.WithFields(log.Fields{
			"address":   a.config.Storage.Address,
			"accessKey": a.config.Storage.AccessKey,
		}).Error("failed to connect to object storage")
		return err
	}
	common.MinioClient = client

	if len(a.config.Storage.Bucket) > 0 {
		common.VideoRecordBucket = a.config.Storage.Bucket
	}
	return nil
}

func (a *Alba) startArchive(date string) error {
	dirs, err := ioutil.ReadDir(a.config.Storage.LiveDir)
	if err != nil {
		return err
	}

	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}

		//log.WithFields(log.Fields{
		//	"id": d.Name(),
		//}).Debugf("found live directory ")

		liveDir := filepath.ToSlash(filepath.Join(a.config.Storage.LiveDir, d.Name())) // live/1/
		if err := a.archive(liveDir, date, d.Name()); err != nil {
			log.Error(err)
			continue
		}

	}
	return nil
}

func (a *Alba) writeLiveFileListToText(liveDir string, files []os.FileInfo, tempDir string) (string, error) {
	var text string
	for _, f := range files {
		//files = append(files, f)
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

func (a *Alba) archive(liveDir, date, streamId string) error {
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

	recordDir := filepath.ToSlash(filepath.Join(a.config.Storage.RecordDir, a.config.Storage.Bucket, streamId, date))
	if err := hippo.EnsureDir(recordDir); err != nil {
		return err
	}
	listFilePath, err := a.writeLiveFileListToText(liveDir, liveFiles, recordDir)
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

func (a *Alba) startScheduler() error {
	a.scheduler.Entries()
	_, err := a.scheduler.AddFunc("10 0 * * *", func() {
		yesterday := time.Now().In(common.Loc).Format(common.DateFormat)

		if err := a.startArchive(yesterday); err != nil {
			log.Error(err)
		}
	})
	return err
}
