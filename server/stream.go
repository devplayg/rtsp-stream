package server

import (
	"fmt"
	"os/exec"
	"strings"
)

type Stream struct {
	Id        string `json:"id"`
	Uri       string `json:"uri"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Recording bool
	Active    bool
	LiveDir   string
	RecDir    string
	CmdType   int
	cmd       *exec.Cmd
}

//
//func NewStream(uri string) *Stream {
//	return &Stream{
//		Uri: uri,
//	}
//}

func (s *Stream) start() error {
	s.Active = true
	return nil
}

func (s *Stream) stop() error {
	err := s.cmd.Process.Kill()
	if strings.Contains(err.Error(), "process already finished") {
		return nil
	}
	if strings.Contains(err.Error(), "signal: killed") {
		return nil
	}
	return err
}

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
