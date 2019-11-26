package streaming

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/minio/highwayhash"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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
	return files, err
}

func MergeLiveVideoFiles(inputFilePath, outputFilePath string) (float32, error) {
	var duration float32
	cmd := exec.Command(
		"ffmpeg",
		"-y",
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
	//log.Debug(cmd.Args)
	var stdOut bytes.Buffer
	cmd.Stdout = &stdOut

	log.Debug(cmd.Args)
	err := cmd.Run()
	if err != nil {
		return duration, err
	}

	if _, err := os.Stat(outputFilePath); os.IsNotExist(err) {
		return duration, errors.New("failed to archive *.ts files: " + stdOut.String())
	}

	durStr, err := GetVideoDuration(outputFilePath)
	if err != nil {
		return duration, errors.New("failed to get duration of file: " + outputFilePath)
	}

	dur, err := strconv.ParseFloat(durStr, 32)
	return float32(dur), nil

	//videoRecordFile := VideoRecord{
	//	Name: filepath.Base(outputFilePath),
	//	Duration: float32(duration),
	//}

	//return &videoRecordFile, err
	//err := s.cmd.Process.Kill()
	//if strings.Contains(err.Error(), "process already finished") {
	//	return nil
	//}
	//if strings.Contains(err.Error(), "signal: killed") {
	//	return nil
	//}
	//return err
}

func GetVideoDuration(path string) (string, error) {

	cmd := exec.Command(
		"ffprobe",
		"-v",
		"error",
		"-show_entries",
		"format=duration",
		"-of",
		"default=noprint_wrappers=1:nokey=1",
		path,
	)
	log.Debug(cmd.Args)

	var stdOut bytes.Buffer
	cmd.Stdout = &stdOut
	err := cmd.Run()
	return strings.TrimSpace(stdOut.String()), err
}

func GenerateLiveVideoFileListToMergeForUseWithFfmpeg(liveVideoFiles []*LiveVideoFile) (*os.File, error) {
	var text string
	for _, f := range liveVideoFiles {
		path := filepath.ToSlash(filepath.Join(f.Dir, f.File.Name()))
		text += fmt.Sprintf("file %s\n", path)
	}
	tempFile, err := ioutil.TempFile("", "stream")
	if err != nil {
		return nil, err
	}
	defer tempFile.Close()
	_, err = tempFile.WriteString(text)
	if err != nil {
		return nil, err
	}

	return tempFile, nil
}

func GetVideRecordBucket(videoRecord *VideoRecord, id int64) []byte {
	t := time.Unix(videoRecord.UnixTime, 0).In(Loc)
	return []byte(fmt.Sprintf("stream-%d-%s", id, t.Format("20060102")))
}

func GetM3u8Header(firstSeq int64, maxTargetDuration float64) string {
	return fmt.Sprintf(`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:%d
#EXT-X-TARGETDURATION:%.0f
`, firstSeq, maxTargetDuration)
}

func GetM3u8Footer() string {
	return "#EXT-X-ENDLIST"
}
