package streaming

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/minio/highwayhash"
	"github.com/minio/minio-go"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
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

	w.Header().Add("Content-Type", ContentTypeJson)
	b, _ := json.Marshal(NewResult(err))
	w.WriteHeader(statusCode)
	w.Write(b)
}

func GetHashString(str string) string {
	hash := highwayhash.Sum128([]byte(str), HashKey)
	return hex.EncodeToString(hash[:])
}

func GetHashFromFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hash, err := highwayhash.New128(HashKey)
	if err != nil {
		return nil, err
	}

	if _, err = io.Copy(hash, file); err != nil {
		return nil, err
	}

	return hash.Sum(nil), nil
}

func GetStreamingCommand(stream *Stream) *exec.Cmd {
	if stream.Protocol == HLS {
		return GetHlsStreamingCommand(stream)
	}

	return GetHlsStreamingCommand(stream)

}

//func checkStreamUri(stream *Stream) error {
//	uri := strings.TrimPrefix(stream.Uri, "rtsp://")
//	stream.Uri = fmt.Sprintf("rtsp://%s:%s@%s", stream.Username, stream.Password, uri)
//
//	return nil
//}

//
//func GetRecentFilesInDir(dir string, after time.Duration) ([]*LiveVideoFile, error) {
//	files := make([]*LiveVideoFile, 0)
//	list, err := ioutil.ReadDir(dir)
//	if err != nil {
//		return nil, err
//	}
//
//	for _, f := range list {
//		if f.IsDir() {
//			continue
//		}
//
//		if time.Since(f.ModTime()) < after {
//			continue
//		}
//
//		if f.Size() < 1 {
//			continue
//		}
//
//		ext := filepath.Ext(f.Name())
//		if ext != ".ts" {
//			continue
//		}
//
//		files = append(files, NewLiveVideoFile(f, ext, dir))
//
//	}
//	return files, err
//}

func MergeLiveVideoFiles(inputFilePath, outputFilePath string) error {
	if err := os.Chdir(filepath.Dir(inputFilePath)); err != nil {
		return nil
	}
	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-f",
		"concat",
		"-safe",
		"0",
		"-i",
		filepath.Base(inputFilePath),
		"-c",
		"copy",
		"-f",
		"ssegment",
		"-segment_list",
		filepath.Base(outputFilePath),
		"-segment_list_flags",
		"+live",
		"-segment_time",
		"10",
		VideoFilePrefix+"%d.ts",
	)
	//output, err := cmd.CombinedOutput()
	//if err != nil {
	//    log.Error(string(output))
	//    return err
	//}
	return cmd.Run()
}

func GetVideoFilesInDir(dir string, prefix string) ([]*VideoFile, error) {
	videoFiles := make([]*VideoFile, 0)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}

		if f.Size() < 1 {
			continue
		}

		ext := filepath.Ext(f.Name())
		if ext != VideoFileExt {
			continue
		}

		if !strings.HasPrefix(f.Name(), prefix) {
			continue
		}

		videoFiles = append(videoFiles, NewVideoFile(f, dir))
	}
	sort.SliceStable(videoFiles, func(i, j int) bool {
		return videoFiles[i].File.ModTime().Before(videoFiles[j].File.ModTime())
	})

	return videoFiles, nil
}

func SendToStorage(bucketName, objectName, path, contentType string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return err
	}

	if len(contentType) < 1 {
		contentType = "application/octet-stream"
	}
	if _, err = MinioClient.PutObject(bucketName, objectName, file, fileStat.Size(), minio.PutObjectOptions{ContentType: contentType}); err != nil {
		return err
	}
	return nil
}

func SendToVirtualStorage(bucketName, objectName, path, contentType string) error {
	return nil
}

func GetVideoFileSeq(name string) (int, error) {
	str := strings.TrimPrefix(filepath.Base(name), VideoFilePrefix)
	str = strings.TrimSuffix(str, VideoFileExt)
	mediaFileSeq, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}

	return mediaFileSeq, nil
}

//
//func GetVideoDuration(path string) (string, error) {
//
//    cmd := exec.Command(
//        "ffprobe",
//        "-v",
//        "error",
//        "-show_entries",
//        "format=duration",
//        "-of",
//        "default=noprint_wrappers=1:nokey=1",
//        path,
//    )
//    log.Debug(cmd.Args)
//
//    var stdOut bytes.Buffer
//    cmd.Stdout = &stdOut
//    err := cmd.Run()
//    return strings.TrimSpace(stdOut.String()), err
//}

//func GenerateLiveVideoFileListToMergeForUseWithFfmpeg(liveVideoFiles []*LiveVideoFile) (*os.File, error) {
//	var text string
//	for _, f := range liveVideoFiles {
//		path := filepath.ToSlash(filepath.Join(f.Dir, f.File.Name()))
//		text += fmt.Sprintf("file %s\n", path)
//	}
//	tempFile, err := ioutil.TempFile("", "stream")
//	if err != nil {
//		return nil, err
//	}
//	defer tempFile.Close()
//	_, err = tempFile.WriteString(text)
//	if err != nil {
//		return nil, err
//	}
//
//	return tempFile, nil
//}

//func GetVideRecordBucket(videoRecord *VideoRecord, id int64) []byte {
//    t := time.Unix(videoRecord.UnixTime, 0).In(Loc)
//    return []byte(fmt.Sprintf("stream-%d-%s", id, t.Format(DateFormat)))
//}
//
//func GetM3u8Header(firstSeq int64, maxTargetDuration float64) string {
//    return fmt.Sprintf(`#EXTM3U
//#EXT-X-VERSION:3
//#EXT-X-MEDIA-SEQUENCE:%d
//#EXT-X-TARGETDURATION:%.0f
//`, firstSeq, maxTargetDuration)
//}
//
//func GetM3u8Footer() string {
//    return "#EXT-X-ENDLIST"
//}
