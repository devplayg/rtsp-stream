package streaming

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/minio/highwayhash"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os/exec"
	"path/filepath"
	"time"
)

var HashKey []byte

func init() {
	data := []byte("this is the key")
	sum := sha256.Sum256(data)
	HashKey = sum[:]
}

func Response(w http.ResponseWriter, err error, statusCode int) {
	if statusCode != http.StatusOK {
		log.Error(err)
	}

	w.Header().Add("Content-Type", ApplicationJson)
	b, _ := json.Marshal(NewResult(err))
	w.WriteHeader(statusCode)
	w.Write(b)
}

func GetHashString(str string) string {
	hash := highwayhash.Sum128([]byte(str), HashKey)
	return hex.EncodeToString(hash[:])
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
		//"-hls_time",
		//"60",
		"-hls_segment_filename",
		fmt.Sprintf("%s/live%%d.ts", stream.LiveDir),
		fmt.Sprintf("%s/index.m3u8", stream.LiveDir),
	)
	return cmd
}

func GetRecentFilesInDir(dir string, after time.Duration) ([]*LiveVideoFile, error) {

	files := make([]*LiveVideoFile, 0)

	//err:= filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
	//	if err != nil {
	//		log.Error(err)
	//		return nil
	//	}
	//
	//	if f.IsDir() {
	//		return nil
	//	}
	//
	//	if f.Size() < 1 {
	//		return nil
	//	}
	//
	//	ext := filepath.Ext(f.Name())
	//	if ext != ".ts" {
	//		return nil
	//	}
	//
	//
	//	files = append(files, NewVideoFile(f, ext, dir ))
	//
	//	return nil
	//})

	//var text string
	list, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, f := range list {
		if f.IsDir() {
			continue
		}

		if time.Since(f.ModTime()) < after {
			continue
		}

		if f.Size() < 1 {
			continue
		}

		ext := filepath.Ext(f.Name())
		if ext != ".ts" {
			continue
		}

		files = append(files, NewLiveVideoFile(f, ext, dir))

	}

	//path := filepath.Join(dir, f.Name())
	//text += fmt.Sprintf("file '%s'\n", path)
	return files, err
}

func ArchiveLiveVideos(inputFilePath, outputFilePath string) error {
	cmd := exec.Command(
		"ffmpeg",
		"-f",
		"concat",
		"-safe",
		"0",
		"-i",
		inputFilePath,
		"-c",
		"copy",
		outputFilePath,
	)

	return cmd.Run()

}
