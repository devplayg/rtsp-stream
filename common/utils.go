package common

import (
	"encoding/binary"
	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
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

func ReadVideoFilesOnDateInDir(dir, date, ext string) ([]os.FileInfo, error) {
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
			//if err := os.Remove(filepath.Join(dir, f.Name())); err != nil {
			//	log.Error(err)
			//}
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

//func ReadVideoFilesInDirNotOnDate(dir, date, ext string) ([]os.FileInfo, error) {
//	list, err := ioutil.ReadDir(dir)
//	if err != nil {
//		return nil, err
//	}
//
//	//var text string
//	files := make([]os.FileInfo, 0)
//	for _, f := range list {
//		if f.IsDir() {
//			continue
//		}
//
//		if !f.Mode().IsRegular() {
//			continue
//		}
//
//		if !strings.HasSuffix(f.Name(), ext) {
//			continue
//		}
//
//		if f.ModTime().In(Loc).Format("20060102") == date {
//			continue
//		}
//
//		files = append(files, f)
//	}
//
//	return files, nil
//}

func RemoveLiveFiles(dir string, files []os.FileInfo) int {
	count := 0
	for _, f := range files {
		if err := os.Remove(filepath.Join(dir, f.Name())); err != nil {
			log.Error(err)
			continue
		}
		count++
	}
	return count
}

func MergeLiveVideoFiles(listFilePath, metaFilePath string, segmentTime int) error {
	inputFile, _ := filepath.Abs(listFilePath)
	//outputFile := filepath.Base(metaFilePath)

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
		//"-segment_list",
		//outputFile,
		"-segment_list_flags",
		"+cache",
		"-segment_time",
		strconv.Itoa(segmentTime),
		VideoFilePrefix+"%d.ts",
	)
	//output, err := cmd.CombinedOutput()
	//if err != nil {
	//   log.Error(string(output))
	//   return []byte{}, err
	//}
	log.Debug(cmd.Args)
	err = cmd.Run()
	if err != nil {
		log.Error(cmd.Args)
	}

	return err
}

func GetDbBucketList(db *bolt.DB, prefix string) ([]string, error) {
	bucketNames := make([]string, 0)
	err := db.View(func(tx *bolt.Tx) error {
		err := tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
			if len(prefix) > 0 {
				if strings.HasPrefix(string(name), prefix) {
					bucketNames = append(bucketNames, string(name))
					return nil
				}
				return nil
			}
			bucketNames = append(bucketNames, string(name))
			return nil
		})
		return err
	})
	return bucketNames, err
}

func CreateDefaultDayRecord(date string, bucketNames []string) map[string]string {
	m := make(map[string]string)
	m["date"] = date
	for _, name := range bucketNames {
		m[name] = ""
	}

	return m
}

func SortDayRecord(dayRecordMap DayRecordMap) []map[string]string {
	keys := make([]string, 0)
	for date, _ := range dayRecordMap {
		keys = append(keys, date)
	}

	sort.Strings(keys)
	values := make([]map[string]string, 0)
	for _, k := range keys {
		values = append(values, dayRecordMap[k])
	}

	return values
}
