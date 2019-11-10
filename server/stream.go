package server

import "os/exec"

type Stream struct {
	Url string
	//Username string
	//Password string
	cmd *exec.Cmd
	//liveDir  string
	//recDir string
	recording bool

	Storage struct {
		Recording string
		Live      string
	}
}

func NewStream(url string) *Stream {
	return &Stream{
		Url: url,
	}
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
