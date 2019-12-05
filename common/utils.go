package common

import (
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
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
