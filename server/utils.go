package server

import (
	"encoding/binary"
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

func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}

func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}
