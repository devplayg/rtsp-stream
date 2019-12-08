package common

import (
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}

func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func ReadVideoFilesInDirOnDate(dir, date, ext string) ([]os.FileInfo, error) {
	list, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	//var text string
	files := make([]os.FileInfo, 0)
	for _, f := range list {
		if f.IsDir() {
			continue
		}

		if !f.Mode().IsRegular() {
			continue
		}

		if f.Size() < 1 {
			if err := os.Remove(filepath.Join(dir, f.Name())); err != nil {
				log.Error(err)
			}
			continue
		}

		if !strings.HasSuffix(f.Name(), ext) {
			continue
		}

		if f.ModTime().In(Loc).Format("20060102") != date {
			continue
		}

		//files = append(files, f)

		//path := filepath.ToSlash(filepath.Join(dir, f.Name()))

		//text += fmt.Sprintf("file '%s'\n", path)
		files = append(files, f)
	}

	return files, nil
}

func RemoveLiveFiles(dir string, files []os.FileInfo) {
	for _, f := range files {
		if err := os.Remove(filepath.Join(dir, f.Name())); err != nil {
			log.Error(err)
		}
	}
}

func MergeLiveVideoFiles(listFilePath, metaFilePath string) error {
	inputFile, _ := filepath.Abs(listFilePath)
	outputFile := filepath.Base(metaFilePath)

	originDir, err := os.Getwd()
	if err != nil {
		return err
	}

	if err := os.Chdir(filepath.Dir(listFilePath)); err != nil {
		return err
	}

	defer os.Chdir(originDir)

	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-f",
		"concat",
		"-safe",
		"0",
		"-i",
		inputFile,
		"-c",
		"copy",
		"-f",
		"ssegment",
		"-segment_list",
		outputFile,
		"-segment_list_flags",
		"+live",
		"-segment_time",
		"30",
		VideoFilePrefix+"%d.ts",
	)
	//output, err := cmd.CombinedOutput()
	//if err != nil {
	//   log.Error(string(output))
	//   return []byte{}, err
	//}
	err = cmd.Run()
	if err != nil {
		log.Error(cmd.Args)
	}

	return err
}
