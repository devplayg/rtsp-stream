package server

import "github.com/pkg/errors"

const (
	Stopped = 0
	Running = 1
)

var (
	StreamBucket = []byte("stream")
	ConfigBucket = []byte("config")
)


var ApplicationJson = "application/json"
var ErrorInvalidUri = errors.New("invalid URI")

type Result struct {
	Error error `json:"error"`
}
