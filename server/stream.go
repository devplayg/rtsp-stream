package server

import (
	"fmt"
	"os/exec"
	"strings"
)

type Stream struct {
	URI      string `json:"uri"`
	Username string
	Password string
	cmd      *exec.Cmd `json:"-"`
	//liveDir  string
	//recDir string
	recording bool
	status    int

	Storage struct {
		Recording string
		Live      string
	} `json:"-"`
}

func NewStream(uri string) *Stream {
	return &Stream{
		URI: uri,
	}
}

func (s *Stream) Start() error {
	s.status = Running
	return nil
}

func (s *Stream) Stop() error {
	err := s.cmd.Process.Kill()
	if strings.Contains(err.Error(), "process already finished") {
		return nil
	}
	if strings.Contains(err.Error(), "signal: killed") {
		return nil
	}
	return err
}

// Stream describes a given host's streaming
//type Stream struct {
//	CMD         *exec.Cmd            `json:"-"`
//	Mux         *sync.RWMutex        `json:"-"`
//	Path        string               `json:"path"`
//	Streak      *hotstreak.Hotstreak `json:"-"`
//	OriginalURI string               `json:"-"`
//	StorePath   string               `json:"-"`
//	KeepFiles   bool                 `json:"-"`
//	Logger      *lumberjack.Logger   `json:"-"`
//}

func generateStreamCmd(uri string, dir string) *exec.Cmd {
	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-fflags",
		"nobuffer",
		"-rtsp_transport",
		"tcp",
		"-i",
		uri,
		"-vsync",
		"0",
		"-copyts",
		"-vcodec",
		"copy",
		"-movflags",
		"frag_keyframe+empty_moov",
		"-an",
		"-hls_flags",
		"append_list",
		"-f",
		"hls",
		"-segment_list_flags",
		"live",
		"-hls_time",
		"1",
		"-hls_list_size",
		"3",
		"-hls_segment_filename",
		fmt.Sprintf("%s/%%d.ts", dir),
		fmt.Sprintf("%s/index.m3u8", dir),
	)
	return cmd
}
