package server

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os/exec"
)

var HashKey []byte

func init() {
	data := []byte("this is the key")
	sum := sha256.Sum256(data)
	HashKey = sum[:]
}

func ResponseError(w http.ResponseWriter, err error, status int) {
	log.Error(err)
	w.Header().Add("Content-Type", ApplicationJson)
	b, _ := json.Marshal(Result{Error: err.Error()})
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

//func checkStreamUri(stream *Stream) error {
//	uri := strings.TrimPrefix(stream.Uri, "rtsp://")
//	stream.Uri = fmt.Sprintf("rtsp://%s:%s@%s", stream.Username, stream.Password, uri)
//
//	return nil
//}

func GenerateStreamCommand(stream *Stream) *exec.Cmd {
	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-fflags",
		"nobuffer",
		"-rtsp_transport",
		"tcp",
		"-i",
		stream.StreamUri(),
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
		fmt.Sprintf("%s/live%%d.ts", stream.LiveDir),
		fmt.Sprintf("%s/index.m3u8", stream.LiveDir),
	)
	return cmd
}
