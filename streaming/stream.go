package streaming

import (
	"context"
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Stream struct {
	Id                 int64         `json:"id"`           // Stream unique ID
	Uri                string        `json:"uri"`          // Stream URL
	Username           string        `json:"username"`     // Stream username
	Password           string        `json:"password"`     // Stream password
	Recording          bool          `json:"recording"`    // Is recording
	Enabled            bool          `json:"enabled"`      // Enabled
	Protocol           int           `json:"protocol"`     // Protocol (HLS, WebM)s
	ProtocolInfo       *ProtocolInfo `json:"protocolInfo"` // Protocol info
	UrlHash            string        `json:"urlHash"`      // URL Hash
	cmd                *exec.Cmd     `json:"-"`            // Command
	liveDir            string        `json:"-"`            // Live video directory
	Status             int           `json:"status"`       // Stream status
	DataRetentionHours int           `json:"dataRetentionHours"`
}

func NewStream() *Stream {
	return &Stream{}
}

func (s *Stream) IsActive() bool {
	if s.cmd == nil {
		return false
	}

	// Check if index file exists
	path := filepath.Join(s.liveDir, s.ProtocolInfo.MetaFileName)
	file, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	// Check if "index.m3u8" has been updated within the last 10 seconds
	if time.Now().Sub(file.ModTime()).Seconds() > 8.0 {
		return false
	}

	// Check if the .ts file is created continuously
	// wondory

	return true
}

func (s *Stream) StreamUri() string {
	uri := strings.TrimPrefix(s.Uri, "rtsp://")
	return fmt.Sprintf("rtsp://%s:%s@%s", s.Username, s.Password, uri)
}

func (s *Stream) WaitUntilStreamingStarts(startedChan chan<- bool, ctx context.Context) {
	count := 1
	for {
		active := s.IsActive()
		log.WithFields(log.Fields{
			"active": active,
			"count":  count,
		}).Debugf("    [stream-%d] is waiting until streaming starts", s.Id)
		if active {
			startedChan <- true
			return
		}
		count++

		select {
		case <-time.After(1 * time.Second):
		case <-ctx.Done():
			return
		}
	}
}

func (s *Stream) start() error {
	s.cmd = GetHlsStreamingCommand(s)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	// Start process
	go func() {
		s.Status = Starting
		err := s.cmd.Run()
		log.WithFields(log.Fields{
			"err": err,
			"pid": GetStreamPid(s),
		}).Debugf("    [stream-%d] process has been terminated", s.Id)
		// s.stop()
		cancel()
	}()

	// Wait until streaming starts
	startedChan := make(chan bool)
	go func() {
		s.WaitUntilStreamingStarts(startedChan, ctx)
	}()

	// Wait signals
	select {
	case <-startedChan:
		log.WithFields(log.Fields{
			"id":  s.Id,
			"pid": GetStreamPid(s),
		}).Debugf("    [stream-%d] stream has been started", s.Id)
		s.Status = Started
		return nil
	case <-ctx.Done():
		msg := "canceled"
		//log.WithFields(log.Fields{
		//    "id": s.Id,
		//}).Debugf("    [stream-%d] %s", s.Id, msg)
		s.stop()
		s.Status = Failed
		return errors.New(msg)
	}

}

func (s *Stream) stop() {
	defer func() {
		s.Status = Stopped
	}()
	if s.cmd == nil || s.cmd.Process == nil {
		return
	}
	//err := s.cmd.Process.Kill()
	err := s.cmd.Process.Signal(os.Kill)
	log.WithFields(log.Fields{
		"uri": s.Uri,
		"err": err,
	}).Debugf("    [stream-%d] has been stopped", s.Id)
	spew.Dump(s.cmd.Process.Signal(os.Kill))

}

//
//func (s *Stream) stop() error {
//	err := s.cmd.Process.Kill()
//	if strings.Contains(err.Error(), "process already finished") {
//		return nil
//	}
//	if strings.Contains(err.Error(), "signal: killed") {
//		return nil
//	}
//	return err
//}

//
//func (p Processor) getHLSFlags() string {
//    if p.keepFiles {
//        return "append_list"
//    }
//    return "delete_segments+append_list"
//}

//
