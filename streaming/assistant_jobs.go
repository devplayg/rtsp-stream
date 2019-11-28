package streaming

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/devplayg/hippo"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (s *Assistant) archiveLiveVideos() error {

	t := time.Now().In(Loc)
	if t.Format(DateFormat) != s.date {
		s.date = t.Format(DateFormat)
		s.lastSentMediaFileSeq = -1
	}
	tempDir := filepath.Join(s.stream.LiveDir, t.Format(DateFormat))
	if err := hippo.EnsureDir(tempDir); err != nil {
		return err
	}

	// Get live video files
	liveVideoFiles, err := s.getLiveVideoFilesToMove(t)
	if err != nil || len(liveVideoFiles) < 1 {
		return err
	}

	// Move live video files to temp directory
	if err := s.moveLiveVideoFilesToTempDir(liveVideoFiles, tempDir); err != nil {
		return err
	}

	// Split video files in temp directory and generate M3U8
	if err := s.generateHlsFiles(tempDir); err != nil {
		return err
	}

	// Send merged video files to object storage
	if err := s.sendVideoFilesToStorage(tempDir, t); err != nil {
		return err
	}

	//// Generate file list of live video files for use with ffmpeg
	//fileList, err := GenerateLiveVideoFileListToMergeForUseWithFfmpeg(liveVideoFiles)
	//if err != nil {
	//	return err
	//}
	//defer os.Remove(fileList.Name())
	//
	//// Merge live video files
	//videoRecord, err := s.mergeLiveVideoFiles(liveVideoFiles, listFile)
	//if err != nil {
	//	return err
	//}

	//err = s.saveVideoRecord(videoRecord)
	//if err != nil {
	//	return err
	//}
	//
	//log.WithFields(log.Fields{
	//	"stream_id": s.stream.Id,
	//	"seq":       videoRecord.Seq,
	//}).Debug("videoRecord has been saved")
	//
	//if err := s.putVideoFileToObjectStorage(videoRecord); err != nil {
	//	return err
	//}
	//defer os.Remove(videoRecord.path)
	//
	//spew.Dump(videoRecord)

	//// update m3u8
	//if err := s.updateM3u8(videoFile); err != nil {
	//	return err
	//}

	//s.removeLiveVideos(liveVideoFiles)
	return nil
}

func (s *Assistant) sendVideoFilesToStorage(tempDir string, t time.Time) error {

	// Get video files
	videoFiles, err := GetVideoFilesInDir(tempDir, VideoFilePrefix)
	if err != nil {
		return err
	}

	if len(videoFiles) < 1 {
		return nil
	}

	lastSentIdx := -1
	lastSentMediaFileSeq := -1
	count := 0
	for idx, f := range videoFiles {
		mediaFileSeq, err := GetVideoFileSeq(f.File.Name())
		if err != nil {
			log.WithFields(log.Fields{
				"name": f.File.Name(),
			}).Debug("invalid video file name")
			continue
		}

		// Skip sent files
		if mediaFileSeq < s.lastSentMediaFileSeq {
			continue
		}

		// Last sent video is needed to check checksum, because it could be changed
		if mediaFileSeq == s.lastSentMediaFileSeq {
			hash, err := GetHashFromFile(filepath.Join(f.dir, f.File.Name()))
			if err != nil {
				log.WithFields(log.Fields{
					"name": f.File.Name(),
				}).Debugf("[%d] failed to get file hash", s.stream.Id)
				return err
			}
			if bytes.Equal(s.lastSentHash, hash) {
				continue
			}
			log.WithFields(log.Fields{
				"name":     f.File.Name(),
				"hash_old": hex.EncodeToString(s.lastSentHash),
				"size_old": s.lastSentSize,
				"hash_now": hex.EncodeToString(hash),
				"size_now": f.File.Size(),
			}).Debugf("    [%d] already sent before, but hash is changed", s.stream.Id)
		}

		// Send video files to storage
		objectName := fmt.Sprintf("%d/%s/%s", s.stream.Id, t.Format(DateFormat), f.File.Name())
		// wondory "SendToVirtualStorage"
		if err := SendToVirtualStorage(VideoRecordBucket, objectName, filepath.Join(f.dir, f.File.Name()), ContentTypeTs); err != nil {
			return err
		}

		log.WithFields(log.Fields{
			"name": f.File.Name(),
			"size": f.File.Size(),
		}).Debugf("    [%d] sent to object", s.stream.Id)

		lastSentMediaFileSeq = mediaFileSeq
		lastSentIdx = idx
		count++
	}

	// Send m3u8 file to storage
	objectName := fmt.Sprintf("%d/%s/%s", s.stream.Id, t.Format(DateFormat), IndexM3u8)
	// wondory
	if err := SendToVirtualStorage(VideoRecordBucket, objectName, filepath.Join(tempDir, IndexM3u8), ContentTypeM3u8); err != nil {
		return err
	}

	hash, err := GetHashFromFile(filepath.Join(videoFiles[lastSentIdx].dir, videoFiles[lastSentIdx].File.Name()))
	if err != nil {
		log.WithFields(log.Fields{
			"name": videoFiles[lastSentIdx].File.Name(),
		}).Debugf("    [%d] failed to get hash", s.stream.Id)
		return err
	}
	s.lastSentMediaFileSeq = lastSentMediaFileSeq
	s.lastSentHash = hash
	s.lastSentSize = videoFiles[lastSentIdx].File.Size()

	transmissionResult := NewTransmissionResult(s.stream.Id, s.lastSentMediaFileSeq, s.lastSentSize, s.lastSentHash, t.Format(DateFormat))
	if err := s.recordTransmission(transmissionResult); err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"id":                   s.stream.Id,
		"lastSentMediaFileSeq": s.lastSentMediaFileSeq,
		"lastSentHash":         hex.EncodeToString(s.lastSentHash),
		"date":                 t.Format(DateFormat),
		"count":                count,
	}).Debug("sent file to object storage")
	return nil
}

func (s *Assistant) recordTransmission(result *TransmissionResult) error {
	return DB.Update(func(tx *bolt.Tx) error {
		data, err := json.Marshal(result)
		if err != nil {
			return err
		}
		bucket := tx.Bucket(TransmissionBucket)
		key := Int64ToBytes(result.StreamId)
		return bucket.Put(key, data)
	})
}

func (s *Assistant) generateHlsFiles(tempDir string) error {
	// Sort ts files by modification time (live9.ts, live10.ts, live11.ts ...)
	liveVideoFiles, err := GetVideoFilesInDir(tempDir, LiveVideoFilePrefix)
	listFile, err := s.generateListFileForUseWithFfmpeg(tempDir, liveVideoFiles)
	if err != nil {
		return err
	}

	dst := filepath.Join(tempDir, "index.m3u8")
	if err := MergeLiveVideoFiles(listFile, dst); err != nil {
		return err
	}

	return nil
}

func (s *Assistant) moveLiveVideoFilesToTempDir(liveVideoFiles []*VideoFile, tempDir string) error {
	for i, f := range liveVideoFiles {
		src := filepath.Join(f.dir, f.File.Name())
		dst := filepath.Join(tempDir, f.File.Name())
		if err := os.Rename(src, dst); err != nil {
			return err
		}
		liveVideoFiles[i].dir = tempDir
		log.WithFields(log.Fields{
			"name": filepath.Base(dst),
		}).Tracef("    [%d] live video file is moved", s.stream.Id)
	}

	return nil
}

//
func (s *Assistant) stop() error {
	// Send remains to storage
	log.WithFields(log.Fields{
		"stream_id": s.stream.Id,
	}).Debug("stream assistant has been stopped")

	return nil
}

func (s *Assistant) getLiveVideoFilesToMove(t time.Time) ([]*VideoFile, error) {
	liveVideoFiles := make([]*VideoFile, 0)

	files, err := ioutil.ReadDir(s.stream.LiveDir)
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

		if !strings.HasPrefix(f.Name(), LiveVideoFilePrefix) {
			continue
		}

		// Read *.ts files older than N seconds of last modification
		if time.Since(f.ModTime()) < (20 * time.Second) {
			continue
		}

		//if f.ModTime().In(Loc).Format(DateFormat) != t.Format(DateFormat) {
		//    continue
		//}

		liveVideoFiles = append(liveVideoFiles, NewVideoFile(f, s.stream.LiveDir))
	}

	// Sort ts files by modification time (live9.ts, live10.ts, live11.ts ...)
	//sort.SliceStable(liveVideoFiles, func(i, j int) bool {
	//    return liveVideoFiles[i].File.ModTime().Before(liveVideoFiles[j].File.ModTime())
	//})

	return liveVideoFiles, nil
}

func (s *Assistant) generateListFileForUseWithFfmpeg(tempDir string, liveVideoFiles []*VideoFile) (string, error) {
	var text string
	for _, f := range liveVideoFiles {
		path := filepath.ToSlash(filepath.Join(f.dir, f.File.Name()))
		text += fmt.Sprintf("file '%s'\n", path)
		//log.WithFields(log.Fields{
		//    "name": f.File.Name(),
		//}).Debug("    - will be merged")
	}

	path := filepath.ToSlash(filepath.Join(tempDir, "list.txt"))
	if err := ioutil.WriteFile(path, []byte(text), 0644); err != nil {
		return "", err
	}

	return path, nil
}

//
//func (s *Assistant) mergeLiveVideoFiles(liveVideoFiles []*LiveVideoFile, fileList *os.File) (*VideoRecord, error) {
//    // Merge live *.ts files to record *.ts files
//    ext := ".ts"
//    videoRecord := NewVideoRecord(liveVideoFiles[0].File.ModTime(), Loc, ext)
//    tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("stream-%d", s.stream.Id))
//    if err := hippo.EnsureDir(tempDir); err != nil {
//        return nil, err
//    }
//    videoRecord.path = filepath.Join(tempDir, videoRecord.Name)
//    //date := liveVideoFiles[0].File.ModTime().In(Loc).Format("20060102")
//    //if err := hippo.EnsureDir(filepath.Join(s.stream.RecDir, date)); err != nil {
//    //	return nil, err
//    //}
//    //outputFilePath := filepath.Join(s.stream.RecDir, date, videoRecord.Name)
//    duration, err := MergeLiveVideoFiles(fileList.Name(), videoRecord.path)
//    log.WithFields(log.Fields{
//        "stream_id":   s.stream.Id,
//        "liveDir":     s.stream.LiveDir,
//        "recDir":      s.stream.RecDir,
//        "listFile":    fileList.Name(),
//        "mergedFiles": len(liveVideoFiles),
//        //"output":          videoRecord,
//        "archivingResult": err,
//        "duration":        duration,
//    }).Debug("assistant merged video files")
//    for _, f := range liveVideoFiles {
//        log.Tracef("    - %s", f.File.Name())
//    }
//    videoRecord.Duration = duration
//    return videoRecord, err
//}

//
//func (s *Assistant) mergeLiveVideoFiles_old(liveVideoFiles []*LiveVideoFile, file *os.File) (*VideoRecord, error) {
//
//	// Merge live *.ts files to record *.ts files
//	ext := ".ts"
//	videoRecord := NewVideoRecord(liveVideoFiles[0].File.ModTime(), Loc, "", ext)
//	date := liveVideoFiles[0].File.ModTime().In(Loc).Format("20060102")
//	if err := hippo.EnsureDir(filepath.Join(s.stream.RecDir, date)); err != nil {
//		return nil, err
//	}
//	outputFilePath := filepath.Join(s.stream.RecDir, date, videoRecord.Name)
//	duration, err := MergeLiveVideoFiles(file.Name(), outputFilePath)
//	log.WithFields(log.Fields{
//		"stream_id":       s.stream.Id,
//		"liveDir":         s.stream.LiveDir,
//		"recDir":          s.stream.RecDir,
//		"list":            file.Name(),
//		"output":          outputFilePath,
//		"archivingResult": err,
//		"duration":        duration,
//	}).Debug("assistant merged video files")
//	for _, f := range liveVideoFiles {
//		log.Debugf("    - %s", f.File.Name())
//	}
//	videoRecord.Duration = duration
//	return videoRecord, err
//}

//func (s *Assistant) putVideoFileToObjectStorage(videoRecord *VideoRecord) error {
//    bucketName := VideoRecordBucket
//    streamId := strconv.FormatInt(s.stream.Id, 10)
//    date := time.Unix(videoRecord.UnixTime, 0).In(Loc).Format("20060102")
//    objectName := filepath.ToSlash(filepath.Join(streamId, date, VideoFilePrefix+strconv.FormatInt(videoRecord.Seq, 10)+".ts"))
//
//    file, err := os.Open(videoRecord.path)
//    if err != nil {
//        return err
//    }
//    defer file.Close()
//
//    fileStat, err := file.Stat()
//    if err != nil {
//        return err
//    }
//
//    _, err = MinioClient.PutObject(bucketName, objectName, file, fileStat.Size(), minio.PutObjectOptions{ContentType: "application/octet-stream"})
//    if err != nil {
//
//        return err
//    }
//
//    //url, err := MinioClient.PresignedGetObject(bucketName, objectName, time.Second * 24 * 60 * 60, nil)
//    //if err != nil {
//    //	return err
//    //}
//    //videoRecord.Url = url.String()
//
//    return nil
//}

//
//func (s *Assistant) archiveLiveVideos_old() error {
//
//	// Get live video files
//	liveVideoFiles, err := s.getLiveVideoFiles()
//	if err != nil || len(liveVideoFiles) < 1 {
//		return err
//	}
//
//	// Generate file list of live video files for use with ffmpeg
//	fileList, err := GenerateLiveVideoFileListToMergeForUseWithFfmpeg(liveVideoFiles)
//	if err != nil {
//		return err
//	}
//	defer os.Remove(fileList.Name())
//
//	// Merge live video files
//	videoFile, err := s.mergeLiveVideoFiles(liveVideoFiles, fileList)
//	if err != nil {
//		return err
//	}
//	//spew.Dump(videoFile)
//
//	if err := s.saveVideoRecord(videoFile); err != nil {
//		return err
//	}
//
//	// update m3u8
//	if err := s.updateM3u8(videoFile); err != nil {
//		return err
//	}
//
//	s.removeLiveVideos(liveVideoFiles)
//	return nil
//}

//func (s *Assistant) updateM3u8(videoRecord *VideoRecord) error {
//	bucketName := GetVideRecordBucket(videoRecord, s.stream.Id)
//	//videoRecords := make([]*VideoRecord, 0)
//	var maxTargetDuration float32
//	m3u8Header := GetM3u8Header()
//	var body string
//
//	err := DB.View(func(tx *bolt.Tx) error {
//		b := tx.Bucket(bucketName)
//		c := b.Cursor()
//		for k, _ := c.First(); k != nil; k, _ = c.Next() {
//			var videoRecord VideoRecord
//			err := json.Unmarshal(k, &videoRecord)
//			if err != nil {
//				log.Error(err)
//				continue
//			}
//
//			//videoRecords = append(videoRecords, &videoRecord)
//			if videoRecord.Duration > maxTargetDuration {
//				maxTargetDuration = videoRecord.Duration
//			}
//			body += fmt.Sprintf("#EXTINF:%.6f,\n", videoRecord.Duration)
//			body += videoRecord.Name + "\n"
//		}
//
//		return nil
//	})
//	if err != nil {
//		return err
//	}
//
//	m3u8Header += fmt.Sprintf("#EXT-X-TARGETDURATION:%.0f\n", math.Ceil(float64(maxTargetDuration)))
//	m3u8Footer := "#EXT-X-ENDLIST"
//	m3u8 := m3u8Header + body + m3u8Footer
//
//	date := time.Unix(videoRecord.UnixTime, 0).In(Loc).Format("20060102")
//	outputFilePath := filepath.Join(s.stream.RecDir, date, "index.m3u8")
//	return ioutil.WriteFile(outputFilePath, []byte(m3u8), 0644)
//}
//
//func (s *Assistant) saveVideoRecord(videoRecord *VideoRecord) error {
//    bucketName := GetVideRecordBucket(videoRecord, s.stream.Id)
//
//    return DB.Update(func(tx *bolt.Tx) error {
//
//        // Create bucket
//        if _, err := tx.CreateBucketIfNotExists(bucketName); err != nil {
//            return err
//        }
//
//        data, err := json.Marshal(videoRecord)
//        if err != nil {
//            return err
//        }
//
//        bucket := tx.Bucket(bucketName)
//        uid, _ := bucket.NextSequence()
//        videoRecord.Seq = int64(uid)
//        return bucket.Put(Int64ToBytes(videoRecord.Seq), data)
//    })
//}

//func (s *Assistant) generateVideoList([]*LiveVideoFile) (string, error) {
//    return "", nil
//}

func (s *Assistant) removeLiveVideos(list []*VideoFile) {
	//for _, f := range list {
	//    path := filepath.Join(f.Dir, f.File.Name())
	//    //os.Rename(path, filepath.Join(s.stream.RecDir, f.File.Name()))
	//    if err := os.Remove(path); err != nil {
	//        log.Error(err)
	//    }
	//}
}

//
//func (a *Assistant) mergeTsFiles(list []string) ([]string {
//    // merge
//
//    // Take snapshot
//
//    return nil
//}
//
//func (a *Assistant) geLiveTsFiles(dur time.Duration) []string {
//
//    // Sort *.ts files
//    return nil
//}

// Boltdb

/*
	streams
		key: id
		val: stream information

	records
		key: id

*/
