package streaming

import (
	"github.com/pkg/errors"
	"os"
)

const (
	Stopped = 0
	Running = 1

	NormalStream = 1
	DateFormat   = "20060102"
)

var (
	// BoltDB buckets
	StreamBucket       = []byte("stream")
	TransmissionBucket = []byte("transmission")
	ConfigBucket       = []byte("config")

	// Minio buckets
	VideoRecordBucket = "record"

	// Content types
	ContentTypeJson = "application/json"
	ContentTypeTs   = "video/mp2t"
	ContentTypeM3u8 = "application/vnd.apple.mpegurl"

	IndexM3u8 = "index.m3u8"
)

var (
	ErrorInvalidUri       = errors.New("invalid URI")
	ErrorDuplicatedStream = errors.New("duplicated stream")
	ErrorStreamNotFound   = errors.New("stream not found")
)

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

type VideoFile struct {
	File os.FileInfo
	//Ext  string
	dir string
}

func NewVideoFile(f os.FileInfo, dir string) *VideoFile {
	return &VideoFile{
		File: f,
		dir:  dir,
	}
}

type TransmissionResult struct {
	StreamId int64
	Seq      int
	Hash     []byte
	Size     int64
	Date     string
}

func NewTransmissionResult(streamId int64, seq int, size int64, hash []byte, date string) *TransmissionResult {
	return &TransmissionResult{
		StreamId: streamId,
		Seq:      seq,
		Size:     size,
		Hash:     hash,
		Date:     date,
	}
}

//
//type VideoRecord struct {
//    Seq      int64   `json:"seq"`
//    Name     string  `json:"nm"`
//    Duration float32 `json:"dur"`
//    UnixTime int64   `json:"unix"`
//    Url      string  `json:"url"`
//    path     string
//}
//
//func NewVideoRecord(t time.Time, loc *time.Location, ext string) *VideoRecord {
//    return &VideoRecord{
//        Name:     t.In(loc).Format("20060102_150405") + ext,
//        UnixTime: t.Unix(),
//    }
//}
