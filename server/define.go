package server

import "github.com/pkg/errors"

const (
	Stopped = 0
	Running = 1
)

var ApplicationJson = "application/json"
var InvalidUriError = errors.New("invalid URI")

type Result struct {
	Error error `json:"error"`
}
