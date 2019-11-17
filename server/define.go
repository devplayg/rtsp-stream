package server

import "github.com/pkg/errors"

const (
	Stopped = 0
	Running = 1

	NormalStream = 1
)

var (
	StreamBucket = []byte("stream")
	ConfigBucket = []byte("config")
)

var ApplicationJson = "application/json"
var ErrorInvalidUri = errors.New("invalid URI")
var ErrorDuplicatedStream = errors.New("duplicated stream")
var ErrorStreamNotFound = errors.New("stream not found")

type Result struct {
	Error string `json:"error"`
}
