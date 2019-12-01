package streaming

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Stream struct {
	Id           int64         `json:"id"`        // Stream unique ID
	Uri          string        `json:"uri"`       // Stream URL
	Username     string        `json:"username"`  // Stream username
	Password     string        `json:"password"`  // Stream password
	Recording    bool          `json:"recording"` // Is recording
	Enabled      bool          `json:"enabled"`
	Active       bool          `json:"active"`   // Is active
	Protocol     int           `json:"protocol"` // FFmpeg command type
	ProtocolInfo *ProtocolInfo `json:"protocolInfo"`
	UrlHash      string        `json:"urlHash"` // URL Hash
	cmd          *exec.Cmd     `json:"-"`       // Command
	liveDir      string        `json:"-"`       // Live video directory
	Status       int           `json:"status"`  // 1:stopped, 2:stopping, 3:starting, 4:started

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

func (s *Stream) WaitUntilStreamingStarts(ch chan<- bool, ctx context.Context) {
	count := 1
	for {
		active := s.IsActive()
		log.WithFields(log.Fields{
			"id":     s.Id,
			"active": active,
			"count":  count,
		}).Debugf("    [stream-%d] wait until streaming starts", s.Id)
		if active {
			ch <- true
			return
		}
		count++

		select {
		case <-time.After(1 * time.Second):
		case <-ctx.Done():
			//log.WithFields(log.Fields{
			//	"id": s.Id,
			//}).Debugf("    [stream-%d] time exceeded. failed to start stream", s.Id)
			return
		}
	}
}

func (s *Stream) start() error {
	s.cmd = GetHlsStreamingCommand(s)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	// Start process
	go func() {
		s.Status = Starting
		err := s.cmd.Run()
		if strings.Contains(err.Error(), "exit status 1") {
			return
		}
		log.WithFields(log.Fields{
			"err": err,
			"pid": s.cmd.Process.Pid,
		}).Debug("run_result")
	}()

	// Wait until streaming starts
	ch := make(chan bool)
	go func() {
		s.WaitUntilStreamingStarts(ch, ctx)
	}()

	// Wait signals
	select {
	case <-ch:
		log.WithFields(log.Fields{
			"id":  s.Id,
			"pid": s.cmd.Process.Pid,
		}).Debugf("    [stream-%d] stream has been started", s.Id)
		s.Status = Started
		return nil
	case <-ctx.Done():
		//log.WithFields(log.Fields{
		//	"id": s.Id,
		//	"pid": s.cmd.Process.Pid,
		//}).Debugf("    [stream-%d] failed to start stream", s.Id)
		msg := "time exceeded"
		log.WithFields(log.Fields{
			"id": s.Id,
		}).Debugf("    [stream-%d] %s", s.Id, msg)
		s.Status = Failed
		s.stop()
		return errors.New(msg)
	}
}

func (s *Stream) stop() error {
	if s.cmd == nil || s.cmd.Process == nil {
		//log.WithFields(log.Fields{
		//	"id": s.Id,
		//}).Debug("streaming is not running")
		return nil
	}

	//err := stream.cmd.Process.Kill()
	return s.cmd.Process.Signal(os.Kill)
	//log.WithFields(log.Fields{
	//	"id":  s.Id,
	//	"pid": &s.cmd.Process.Pid,
	//}).Error("killed stream process: ", err)

	//return nil
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
