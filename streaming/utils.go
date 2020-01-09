package streaming

import (
	"encoding/json"
	"errors"
	"github.com/devplayg/rtsp-stream/common"
	"github.com/gorilla/mux"
	"github.com/minio/highwayhash"
	"github.com/minio/minio-go"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func GetHlsStreamingCommand(stream *Stream) *exec.Cmd {
	return exec.Command(
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
		stream.liveDir+"/"+stream.ProtocolInfo.LiveFilePrefix+"%d.ts",
		stream.liveDir+"/"+stream.ProtocolInfo.MetaFileName,
	)
	//output, err := cmd.CombinedOutput()
	//if err != nil {
	//    log.Error(string(output))
	//    return err
	//}
}

func GetHashFromFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hash, err := highwayhash.New128(common.HashKey)
	if err != nil {
		return nil, err
	}

	if _, err = io.Copy(hash, file); err != nil {
		return nil, err
	}

	return hash.Sum(nil), nil
}

func GetStreamPid(stream *Stream) int {
	if stream.Cmd == nil || stream.Cmd.Process == nil {
		return 0
	}
	return stream.Cmd.Process.Pid
}

func GetDirSize(dir string) (int64, error) {
	var size int64
	err := filepath.Walk(dir, func(path string, file os.FileInfo, err error) error {
		if file.IsDir() {
			return nil
		}
		if !file.Mode().IsRegular() {
			return nil
		}

		size += file.Size()
		return nil
	})

	return size, err
}

func ParseAndGetStream(body io.Reader) (*Stream, error) {
	stream := NewStream()
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(data, stream); err != nil {
		return nil, err
	}

	stream.Uri = strings.TrimSpace(stream.Uri)
	if _, err := url.Parse(stream.Uri); err != nil {
		return nil, common.ErrorInvalidUri
	}

	return stream, nil
}

func ParseAndGetStreamId(r *http.Request) (int64, error) {
	vars := mux.Vars(r)
	if len(vars["id"]) < 1 {
		return 0, errors.New("empty stream id")
	}

	streamId, _ := strconv.ParseInt(vars["id"], 10, 64)
	return streamId, nil
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
	if _, err = common.MinioClient.PutObject(bucketName, objectName, file, fileStat.Size(), minio.PutObjectOptions{ContentType: contentType}); err != nil {
		return err
	}
	return nil
}

func GetVideoFileSeq(name string) (int, error) {
	str := strings.TrimPrefix(filepath.Base(name), common.VideoFilePrefix)
	str = strings.TrimSuffix(str, common.VideoFileExt)
	mediaFileSeq, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}

	return mediaFileSeq, nil
}

func GetVideoFilesInDir(dir string, prefix string) ([]*common.VideoFile, error) {
	videoFiles := make([]*common.VideoFile, 0)
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
		if ext != common.VideoFileExt {
			continue
		}

		if !strings.HasPrefix(f.Name(), prefix) {
			continue
		}

		videoFiles = append(videoFiles, common.NewVideoFile(f, dir))
	}
	sort.SliceStable(videoFiles, func(i, j int) bool {
		return videoFiles[i].File.ModTime().Before(videoFiles[j].File.ModTime())
	})

	return videoFiles, nil
}
