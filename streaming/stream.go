package streaming

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/devplayg/rtsp-stream/common"
	"github.com/grafov/m3u8"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Stream struct {
	Id                 int64                `json:"id"`           // Stream unique ID
	Uri                string               `json:"uri"`          // Stream URL
	Username           string               `json:"username"`     // Stream username
	Password           string               `json:"password"`     // Stream password
	Recording          bool                 `json:"recording"`    // Is recording
	Enabled            bool                 `json:"enabled"`      // Enabled
	Protocol           int                  `json:"protocol"`     // Protocol (HLS, WebM)s
	ProtocolInfo       *common.ProtocolInfo `json:"protocolInfo"` // Protocol info
	UrlHash            string               `json:"urlHash"`      // URL Hash
	Cmd                *exec.Cmd            `json:"-"`            // Command
	liveDir            string               `json:"-"`            // Live video directory
	Status             int                  `json:"status"`       // Stream status
	DataRetentionHours int                  `json:"dataRetentionHours"`
	Pid                int                  `json:"pid"`
	LastStreamUpdated  time.Time            `json:"lastStreamUpdated"`
	MaxStreamSeqId     int64                `json:"maxStreamSeqId"`
	Created            int64                `json:"created"`
	DB                 *bolt.DB             `json:"-"`
	LastAttemptTime    time.Time            `json:"-"`
	assistant          *Assistant
	ctx                context.Context
	cancel             context.CancelFunc
	// waitTimeUntilStreamStarts time.Duration
}

func NewStream() *Stream {
	return &Stream{}
}

func (s *Stream) GetStatus() (bool, time.Time, float64) {
	active := false
	lastStreamUpdated := time.Time{}

	if s.Cmd == nil || s.Cmd.Process == nil {
		return active, lastStreamUpdated, 0
	}

	// Check if "index.m3u8" has been updated within the last 8 seconds
	path := filepath.Join(s.liveDir, s.ProtocolInfo.MetaFileName)
	//absPath, _ := filepath.Abs(path)
	//pwd, _ := os.Getwd()
	file, err := os.Stat(path)
	var diff float64
	if !os.IsNotExist(err) {
		lastStreamUpdated = file.ModTime()
		diff = time.Now().Sub(file.ModTime()).Seconds()
		if diff <= 12.0 {
			active = true
		}
	}
	//log.WithFields(log.Fields{
	//	"path":    path,
	//	"absPath": absPath,
	//	"pwd":     pwd,
	//	"diff":    diff,
	//}).Debug("status check")

	return active, lastStreamUpdated, diff
}

func (s *Stream) IsActive() bool {
	active, _, _ := s.GetStatus()
	return active
	//if s.cmd == nil || s.cmd.Process == nil {
	//    return false
	//}
	//
	//// Check if index file exists
	//path := filepath.Join(s.liveDir, s.ProtocolInfo.MetaFileName)
	//file, err := os.Stat(path)
	//if os.IsNotExist(err) {
	//    return false
	//}
	//
	//// Check if "index.m3u8" has been updated within the last 8 seconds
	//since := time.Now().Sub(file.ModTime()).Seconds()
	////log.Debugf("    [stream-%d] is active? %3.1f", s.Id, since)
	//if since > 8.0 {
	//    return false
	//}
	//
	//// Check if the .ts file is created continuously
	//// wondory
	//
	//return true
}

func (s *Stream) StreamUri() string {
	uri := strings.TrimPrefix(s.Uri, "rtsp://")
	return fmt.Sprintf("rtsp://%s:%s@%s", s.Username, s.Password, uri)
}

func (s *Stream) WaitUntilStreamingStarts(startedChan chan<- int, ctx context.Context) {
	count := 1
	for {
		if s.IsActive() {
			startedChan <- count

			// Assistant start
			s.assistant = NewAssistant(s)
			s.assistant.start()
			return
		}
		log.WithFields(log.Fields{
			"count": count,
		}).Debugf("    [stream-%d] waiting until streaming starts", s.Id)
		count++

		select {
		case <-time.After(1 * time.Second):
		case <-ctx.Done():
			return
		}
	}
}

func (s *Stream) Start() (int, error) {
	s.LastAttemptTime = time.Now().In(common.Loc)
	s.Cmd = GetHlsStreamingCommand(s)
	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Second)
	go func() {
		// After finishing, you need to do some post-processing
		defer func() {
			if s.assistant != nil {
				s.assistant.stop()
			}
			s.cancel()
			//metaFilePath := filepath.Join(s.liveDir, s.ProtocolInfo.MetaFileName)
			//os.Remove(metaFilePath)
			s.Status = common.Stopped
		}()
		err := s.Cmd.Run()
		log.WithFields(log.Fields{
			"err": err,
			"pid": GetStreamPid(s),
		}).Debugf("    [stream-%d] process has been terminated", s.Id)
		//s.cmd = nil
	}()

	// Wait until streaming starts
	startedChan := make(chan int)
	go func() {
		s.WaitUntilStreamingStarts(startedChan, s.ctx)
	}()

	// Wait signals
	select {
	case count := <-startedChan:
		s.Status = common.Started
		return count, nil
	case <-s.ctx.Done():
		s.Stop()
		s.Status = common.Failed
		return 0, errors.New("failed or canceled")
	}
}

func (s *Stream) Stop() {
	if s.Cmd == nil || s.Cmd.Process == nil {
		return
	}
	err := s.Cmd.Process.Kill()
	log.WithFields(log.Fields{
		"uri":    s.Uri,
		"result": err,
	}).Infof("    [stream-%d] process has been stopped", s.Id)
}

func (s *Stream) makeM3u8Tags(segments []*common.Segment) string {
	size := uint(len(segments))
	playlist, _ := m3u8.NewMediaPlaylist(size, size)
	defer playlist.Close()

	for _, seg := range segments {
		err := playlist.Append(seg.URI, seg.Duration, "")
		playlist.SetDiscontinuity()
		if err != nil {
			log.Error(err)
		}
	}
	if len(segments) > 0 {
		playlist.SeqNo = uint64(segments[0].SeqId)
	}
	//log.WithFields(log.Fields{
	//	"playSeqNo": playlist.SeqNo,
	//	"len(seg)":  len(segments),
	//}).Debug("test1")
	//playlist.MediaType = m3u8.VOD
	//playlist.SetVersion(4)
	playlist.Close()
	return playlist.Encode().String()
}

func (s *Stream) GetM3u8Tags(date string) (string, error) {
	segments, err := s.getM3u8Segments(date)
	if err != nil {
		return "", err
	}
	tags := s.makeM3u8Tags(segments)
	return tags, nil
}

func (s *Stream) getM3u8Segments(date string) ([]*common.Segment, error) {
	segments := make([]*common.Segment, 0)
	err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(date))
		if b == nil {
			return nil
		}
		return b.ForEach(func(k, v []byte) error {
			var seg common.Segment
			if err := json.Unmarshal(v, &seg); err != nil {
				return err
			}
			segments = append(segments, &seg)
			return nil
		})
	})
	return segments, err
}

func (s *Stream) M3u8BucketExists(date string) bool {
	exist := false
	_ = s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(date))
		if b != nil {
			exist = true
		}
		return nil
	})
	return exist
}

func (s *Stream) GetDBFileName() string {
	return "stream-" + strconv.FormatInt(s.Id, 10) + ".db"
}

func (s *Stream) GetLiveDir() string {
	return s.liveDir
}

func (s *Stream) SetLiveDir(dir string) {
	s.liveDir = dir
}
