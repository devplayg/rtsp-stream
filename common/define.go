package common

import (
	"crypto/sha256"
	"github.com/minio/minio-go"
	"github.com/pkg/errors"
	"os"
	"time"
)

var (
	Loc         *time.Location
	MinioClient *minio.Client
	HashKey     []byte
)

func init() {
	data := []byte("this is the key")
	sum := sha256.Sum256(data)
	HashKey = sum[:]
}

const (
	// Status
	Failed   = -1
	Stopped  = 1
	Stopping = 2
	Starting = 3
	Started  = 4

	// Protocol types
	HLS  = 1
	WEBM = 2

	DateFormat = "20060102"

	// Content types
	ContentTypeJson        = "application/json"
	ContentTypeTs          = "video/MP2T"
	ContentTypeM3u8        = "application/x-mpegURL"
	ContentTypeOctetStream = "application/octet-stream"
	//ContentTypeM3u8 = "application/vnd.apple.mpegurl"

	LiveBucketName = "live"

	VideoFilePrefix     = "media"
	LiveVideoFilePrefix = "live"
	VideoFileExt        = ".ts"

	LiveM3u8FileName = "index.m3u8"
)

var (
	// BoltDB buckets
	StreamBucket      = []byte("stream")
	ConfigBucket      = []byte("config")
	VideoBucketPrefix = "video-"
	//TransmissionBucket = []byte("transmission")
	//ConfigBucket       = []byte("config")

	// MinIO buckets
	VideoRecordBucket = "record"
	//IndexM3u8         = "index.m3u8"
	LastArchivingDateKey = []byte("lastRecordingDate")
)

var (
	ErrorInvalidUri       = errors.New("invalid URI")
	ErrorDuplicatedStream = errors.New("duplicated stream")
	ErrorInvalidStream    = errors.New("invalid stream")
	ErrorStreamNotFound   = errors.New("stream not found")
)

type StreamKey struct {
	Id int64 `json:"id"`
}

func NewStreamKey(id int64) *StreamKey {
	return &StreamKey{Id: id}
}

func (s *StreamKey) Marshal() []byte {
	// b, _ := json.Marshal(s)
	return Int64ToBytes(s.Id)
}

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
	dir  string
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

type ProtocolInfo struct {
	Protocol        int    `json:"protocol"`
	MetaFileName    string `json:"metaFileName"`
	LiveFilePrefix  string `json:"liveFilePrefix"`
	VideoFilePrefix string `json:"videoFilePrefix"`
}

func NewProtocolInfo(protocol int) *ProtocolInfo {
	//if protocol == HLS {
	//    return &ProtocolInfo{
	//        MetaFileName:    LiveM3u8FileName,
	//        LiveFilePrefix:  LiveVideoFilePrefix,
	//        VideoFilePrefix: VideoFilePrefix,
	//    }
	//}
	return &ProtocolInfo{
		Protocol:        protocol,
		MetaFileName:    LiveM3u8FileName,
		LiveFilePrefix:  LiveVideoFilePrefix,
		VideoFilePrefix: VideoFilePrefix,
	}
}

type Segment struct {
	SeqId    int64   `json:"id"`
	Duration float64 `json:"d"`
	URI      string  `json:"uri"`
	UnixTime int64   `json:"t"`
	Data     []byte  `json:"-"`
	Date     string  `json:"date"`
}

func NewSegment(seqId int64, duration float64, uri string, modTime time.Time) *Segment {
	return &Segment{
		SeqId:    seqId,
		Duration: duration,
		URI:      uri,
		UnixTime: modTime.Unix(),
		Date:     modTime.Format(DateFormat),
	}
}

type Record struct {
	Id int64
}

type DayRecordMap map[string]map[string]string // rename to dailyVideoMap

type TplGlobalVar struct {
	Common   map[string]interface{}
	Contents interface{}
}

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
