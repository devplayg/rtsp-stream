package store

import (
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/devplayg/hippo"
	"github.com/devplayg/rtsp-stream/common"
	"github.com/minio/minio-go"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"
)

type Alba struct {
	engine *hippo.Engine
	config *common.Config
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

	date := "20191204"
	if err := a.startArchive(date); err != nil {
		log.Error(err)
	}
	log.Debug("")

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

		dir := filepath.ToSlash(filepath.Join(a.config.Storage.LiveDir, d.Name())) // live/1/
		if err := a.archive(dir, date, d.Name()); err != nil {
			log.Error(err)
			continue
		}

	}
	return nil
}

func (a *Alba) writeLiveFileListToText(dir string, files []os.FileInfo, tempDir string) (string, error) {
	var text string
	for _, f := range files {
		//files = append(files, f)
		path := filepath.ToSlash(filepath.Join(dir, f.Name()))
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
		// nothing to do
		return nil
	}

	//tempDir, err := ioutil.TempDir("c:/temp", "video_"+date+"_")
	//if err != nil {
	//    return err
	//}

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
	//
	//
	//t := time.Now()
	//err = MergeLiveVideoFiles(listFilePath, filepath.Join(recordDir, common.LiveM3u8FileName))
	//
	//log.WithFields(log.Fields{
	//    "duration": time.Since(t).Seconds(),
	//    "err": err,
	//}).Debug("merged")
	//
	//if err != nil {
	//    return err
	//}

	//spew.Dump(f.Name())

	//listFilePath := filepath.Join(tempDir, Meta)
	//MergeLiveVideoFiles()

	//if err := common.CreateVideoFileList("list.txt", files, dir, tempDir); err != nil {
	//    return err
	//}
	spew.Dump(listFilePath)

	return nil
}

func MergeLiveVideoFiles(listFilePath, metaFilePath string) error {
	if err := os.Chdir(filepath.Dir(listFilePath)); err != nil {
		return nil
	}
	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-f",
		"concat",
		"-safe",
		"0",
		"-i",
		listFilePath,
		"-c",
		"copy",
		"-f",
		"ssegment",
		"-segment_list",
		metaFilePath,
		"-segment_list_flags",
		"+live",
		"-segment_time",
		"30",
		common.VideoFilePrefix+"%d.ts",
	)
	//output, err := cmd.CombinedOutput()
	//if err != nil {
	//    log.Error(string(output))
	//    return err
	//}
	return cmd.Run()
}
