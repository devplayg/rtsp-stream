package server

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Stream struct {
	Id        int64     `json:"id"`
	Uri       string    `json:"uri"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Recording bool      `json:"recording"`
	Active    bool      `json:"active"`
	LiveDir   string    `json:"-"`
	RecDir    string    `json:"-"`
	Hash      string    `json:"hash"`
	CmdType   int       `json:"-"`
	cmd       *exec.Cmd `json:"-"`
}

//func NewStream(uri string) *Stream {
//	return &Stream{
//		Uri: uri,
//	}
//}
//
//func (s *Stream) start() error {
//	//s.Active = true
//	return nil
//}

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
	if time.Now().Sub(file.ModTime()).Seconds() > 10.0 {
		return false
	}

	// Check if the .ts file is created continuously
	// devplayg

	return true
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

func (s *Stream) StreamUri() string {
	uri := strings.TrimPrefix(s.Uri, "rtsp://")
	return fmt.Sprintf("rtsp://%s:%s@%s", s.Username, s.Password, uri)
}

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
