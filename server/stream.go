package server

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Stream struct {
	Id        int64     `json:"id"`        // Stream unique ID
	Uri       string    `json:"uri"`       // Stream URL
	Username  string    `json:"username"`  // Stream username
	Password  string    `json:"password"`  // Stream password
	Recording bool      `json:"recording"` // Is recording
	Active    bool      `json:"active"`    // Is active
	LiveDir   string    `json:"-"`         // Live video directory
	RecDir    string    `json:"-"`         // Recording directory
	Hash      string    `json:"hash"`      // URL Hash
	CmdType   int       `json:"cmdType"`   // FFmpeg command type
	cmd       *exec.Cmd `json:"-"`         // Command
	// assistant *Assistant `json:"-"`         // Stream assistant
}

func (s *Stream) IsActive() bool {
	if s.cmd == nil {
		return false
	}

	// Check if index file exists
	path := filepath.Join(s.LiveDir, "index.m3u8")
	file, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	// Check if "index.m3u8" has been updated within the last 10 seconds
	if time.Now().Sub(file.ModTime()).Seconds() > 8.0 {
		return false
	}

	// Check if the .ts file is created continuously
	// devplayg

	return true
}

func (s *Stream) StreamUri() string {
	uri := strings.TrimPrefix(s.Uri, "rtsp://")
	return fmt.Sprintf("rtsp://%s:%s@%s", s.Username, s.Password, uri)
}

//func NewStream(uri string) *Stream {
//	return &Stream{
//		Uri: uri,
//	}
//}
//
func (s *Stream) start() error {
	// ctx, cancel := context.WithCancel(context.Background())
	//done := make(chan bool)

	// Start process
	go func() {
		s.run()
		log.WithFields(log.Fields{
			"id": s.Id,
		}).Debug("streaming job is done")
	}()
	log.WithFields(log.Fields{
		"id":      s.Id,
		"uri":     s.Uri,
		"liveDir": s.LiveDir,
		"recDir":  s.RecDir,
	}).Debug("streaming has been started")

	//select {
	//case result := <-done:
	//    return result, nil
	//case <-ctx.Done():
	//    return "Fail", ctx.Err()
	//}

	return nil
}

func (s *Stream) run() error {
	//defer func() {
	//    done <- true
	//}()

	assistant := NewAssistant(s)
	assistant.start()
	// defer s.assistant.stop()
	err := s.cmd.Run()
	if err != nil {
		log.WithFields(log.Fields{
			"id":      s.Id,
			"uri":     s.Uri,
			"liveDir": s.LiveDir,
			"recDir":  s.RecDir,
		}).Debug("streaming has been stopped: ", err)
		return err
	}

	return nil
}

func (s *Stream) stop() error {
	if s.cmd == nil {
		return nil
	}

	//err := stream.cmd.Process.Kill()
	err := s.cmd.Process.Signal(os.Kill)
	log.WithFields(log.Fields{
		"id":      s.Id,
		"uri":     s.Uri,
		"liveDir": s.LiveDir,
		"recDir":  s.RecDir,
	}).Error("result of stopping streaming: ", err)

	return nil
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
//func (m Manager) WaitForStream(path string) chan bool {
//	var once sync.Once
//	streamResolved := make(chan bool, 1)
//
//	// Start scanning for the given file
//	go func() {
//		for {
//			_, err := os.Open(path)
//			if err != nil {
//				<-time.After(25 * time.Millisecond)
//				continue
//			}
//			once.Do(func() { streamResolved <- true })
//			return
//		}
//	}()
//
//	// Start the timeout phase for the restarted stream
//	go func() {
//		<-time.After(m.timeout)
//		once.Do(func() {
//			logrus.Error(fmt.Errorf("%s timed out while waiting for file creation in manager start", path))
//			streamResolved <- false
//		})
//	}()
//
//	return streamResolved
//}
