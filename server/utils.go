package server

import (
    "encoding/json"
    log "github.com/sirupsen/logrus"
    "net/http"
)

func ResponseError(w http.ResponseWriter, err error, status int) {
    log.Error(err)
    w.Header().Add("Content-Type", ApplicationJson)
    b, _ := json.Marshal(Result{Error: err})
    w.WriteHeader(status)
    w.Write(b)
}

