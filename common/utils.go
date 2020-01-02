package common

import (
	"bytes"
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

func CreateStreamKey(id int64) []byte {
	return Int64ToBytes(id)
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

//func GetDbBuckets(db *bolt.DB, prefix string) ([][]byte, error) {
//	buckets := make([][]byte, 0)
//	err := db.View(func(tx *bolt.Tx) error {
//		err := tx.ForEach(func(b []byte, _ *bolt.Bucket) error {
//			if len(prefix) > 0 {
//				if strings.HasPrefix(string(b), prefix) {
//					buckets = append(buckets, b)
//					return nil
//				}
//				return nil
//			}
//			buckets = append(buckets, b)
//			return nil
//		})
//		return err
//	})
//	return buckets, err
//}

func GetVideoRecordHistory(db *bolt.DB) (map[string]map[string]bool, map[string]bool, error) {
	prefix := []byte("video-")
	videoMap := make(map[string]map[string]bool) // videoName / date / bool
	dateMap := make(map[string]bool)             // date / bool
	err := db.View(func(tx *bolt.Tx) error {
		err := tx.ForEach(func(bucketName []byte, b *bolt.Bucket) error {
			if !bytes.HasPrefix(bucketName, prefix) {
				return nil
			}

			b.ForEach(func(date, _ []byte) error {
				if _, ok := videoMap[string(bucketName)]; !ok {
					videoMap[string(bucketName)] = make(map[string]bool)
				}
				videoMap[string(bucketName)][string(date)] = true
				dateMap[string(date)] = true
				return nil
			})
			return nil
		})
		return err
	})
	return videoMap, dateMap, err
}

func CreateDefaultDayRecord(date string, videoNames []string) map[string]interface{} {
	m := make(map[string]interface{})
	m["date"] = date
	for _, name := range videoNames {
		m[name] = 0
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
