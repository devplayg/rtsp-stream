package server

import (
	"github.com/devplayg/rtsp-stream/common"
	"os/exec"
	"path/filepath"
	"strconv"
)

func MergeLiveVideoFiles(listFilePath, metaFilePath string, segmentTime int) error {
	//inputFile, _ := filepath.Abs(listFilePath)
	//outputFile := filepath.Base(metaFilePath)

	//originDir, err := os.Getwd()
	//if err != nil {
	//	return err
	//}

	// /data/record/1/20191217/list.txt -c copy -f ssegment -segment_list /data/record/1/20191217/index.m3u8
	// -segment_list_flags +cache -segment_time 30 /data/record/1/20191217/media%d.ts

	//if err := os.Chdir(filepath.Dir(listFilePath)); err != nil {
	//	return err
	//}
	//defer os.Chdir(originDir)

	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-f",
		"concat",
		"-safe",
		"0",
		"-i",
		listFilePath,
		"-c",
		"copy",
		"-f",
		"ssegment",
		"-segment_list",
		metaFilePath,
		"-segment_list_flags",
		"+cache",
		"-segment_time",
		strconv.Itoa(segmentTime),
		filepath.Join(filepath.Dir(metaFilePath), common.VideoFilePrefix+"%d.ts"),
	)
	//output, err := cmd.CombinedOutput()
	//if err != nil {
	//   log.Error(string(output))
	//   return []byte{}, err
	//}
	//log.WithFields(log.Fields{
	//	"name": "MergeLiveVideoFiles",
	//}).Debug(cmd.Args)
	err := cmd.Run()
	//if err != nil {
	//	log.Error(cmd.Args)
	//}

	return err
}
