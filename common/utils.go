package common

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/boltdb/bolt"
	"github.com/minio/highwayhash"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"sort"
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

func DetectContentType(ext string) string {
	ctype := mime.TypeByExtension(filepath.Ext(ext))
	if ctype == "" {
		return ContentTypeOctetStream
	}
	return ctype
}

func GetHashString(str string) string {
	hash := highwayhash.Sum128([]byte(str), HashKey)
	return hex.EncodeToString(hash[:])
}
