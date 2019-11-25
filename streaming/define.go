package streaming

import (
	"github.com/pkg/errors"
	"os"
	"time"
)

const (
	Stopped = 0
	Running = 1

	NormalStream = 1
)

var (
	StreamBucket = []byte("stream")
	ConfigBucket = []byte("config")
)

var ApplicationJson = "application/json"
var ErrorInvalidUri = errors.New("invalid URI")
var ErrorDuplicatedStream = errors.New("duplicated stream")
var ErrorStreamNotFound = errors.New("stream not found")

type Result struct {
	Error string `json:"error"`
}

func NewResult(err error) *Result {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}

	return &Result{
		Error: errMsg,
	}
}

type LiveVideoFile struct {
	File os.FileInfo
	Ext  string
	Dir  string
}

func NewLiveVideoFile(f os.FileInfo, ext, dir string) *LiveVideoFile {
	return &LiveVideoFile{
		File: f,
		Ext:  ext,
		Dir:  dir,
	}
}

type VideoRecord struct {
	Name     string  `json:"nm"`
	Duration float32 `json:"dur"`
	UnixTime int64   `json:"unix"`
}

func NewVideoRecord(t time.Time, loc *time.Location, ext string) *VideoRecord {
	return &VideoRecord{
		Name:     t.In(loc).Format("20060102_150405") + ext,
		UnixTime: t.Unix(),
	}
}
