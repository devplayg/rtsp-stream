package streaming

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/devplayg/rtsp-stream/common"
	"github.com/grafov/m3u8"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Assistant struct {
	mpu8CaptureInterval   time.Duration
	healthCheckInterval   time.Duration
	capturingM3u8Interval time.Duration
	stream                *Stream
	ctx                   context.Context
	cancel                context.CancelFunc
}

func NewAssistant(stream *Stream) *Assistant {
	assistant := &Assistant{
		mpu8CaptureInterval: 1500 * time.Millisecond,
		healthCheckInterval: 4 * time.Second,
		stream:              stream,
	}
	ctx, cancel := context.WithCancel(context.Background())
	assistant.ctx = ctx
	assistant.cancel = cancel

	return assistant
}

func (s *Assistant) init() error {
	return nil
}

func (s *Assistant) start() error {
	go s.startCapturingLiveM3u8(3)
	//go s.startMergingVideoFiles()
	log.WithFields(log.Fields{}).Debugf("    [assistant-%d] has been started", s.stream.Id)

	return nil
}

func (s *Assistant) startCapturingLiveM3u8(size int) {
	for {
		if s.stream.Status == common.Started {
			if err := s.captureLiveM3u8(size); err != nil {
				log.Error(err)
			}
		}
		//segs := getM3u8Segments(stream, "")
		//tags := makeM3u8Tags(stream, segs)

		select {
		case <-time.After(s.mpu8CaptureInterval):
		case <-s.ctx.Done():
			log.WithFields(log.Fields{}).Debugf("    [assistant-%d] capturing m3u8 has been stopped", s.stream.Id)
			return
		}

	}
}

func (s *Assistant) captureLiveM3u8(size int) error {
	playlist, err := s.readLiveM3u8(size)
	if err != nil {
		return err
	}

	if playlist == nil {
		log.Warn("m3u8 length is zero")
		return nil
	}

	segments, maxSeqId := s.generateSegments(playlist)
	if len(segments) < 1 {
		return nil
	}
	s.stream.MaxStreamSeqId = maxSeqId
	if err := s.saveSegments(segments); err != nil {
		return nil
	}

	log.WithFields(log.Fields{
		"count":     len(segments),
		"lastSeqId": maxSeqId,
	}).Debugf("    [stream-%d] read m3u8", s.stream.Id)

	return nil
}

func (s *Assistant) saveSegments(segments map[int64]*common.Segment) error {
	return s.stream.db.Update(func(tx *bolt.Tx) error {
		for seqId, seg := range segments {
			bucketName := []byte(time.Unix(seg.UnixTime, 0).In(common.Loc).Format(common.DateFormat))
			bucket, err := tx.CreateBucketIfNotExists(bucketName)
			if err != nil {
				return err
			}
			log.WithFields(log.Fields{
				"bucket":  string(bucketName),
				"key":     seqId,
				"segId":   seg.SeqId,
				"segUri":  seg.URI,
				"segTime": seg.UnixTime,
			}).Debug("saved a segment")
			if err = bucket.Put(common.Int64ToBytes(seqId), seg.Data); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Assistant) generateSegments(playlist *m3u8.MediaPlaylist) (map[int64]*common.Segment, int64) {
	m := make(map[int64]*common.Segment)
	var maxSeqId uint64
	for _, seg := range playlist.Segments {
		if seg == nil {
			continue
		}

		file, err := os.Stat(filepath.Join(s.stream.liveDir, seg.URI))
		if err != nil {
			log.Error(err)
			continue
		}
		if file.Size() < 1 {
			log.Warnf("    [stream-%d] file size is zero: %s", s.stream.Id, file.Name())
			continue
		}

		if seg.SeqId > maxSeqId {
			maxSeqId = seg.SeqId
		}

		str := strings.TrimSuffix(strings.TrimPrefix(seg.URI, common.LiveVideoFilePrefix), common.VideoFileExt)
		seqId, _ := strconv.ParseInt(str, 10, 16)
		segment := common.NewSegment(seqId, seg.Duration, seg.URI, file.ModTime().Unix())
		data, _ := json.Marshal(common.NewSegment(seqId, seg.Duration, seg.URI, file.ModTime().Unix()))
		segment.Data = data
		m[seqId] = segment
	}
	return m, int64(maxSeqId)
}

func (s *Assistant) readLiveM3u8(size int) (*m3u8.MediaPlaylist, error) {
	path := filepath.Join(s.stream.liveDir, s.stream.ProtocolInfo.MetaFileName)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	playlist, err := m3u8.NewMediaPlaylist(3, uint(size))
	playlist.DecodeFrom(bufio.NewReader(file), true)

	return playlist, nil
}

func (s *Assistant) stop() error {
	s.cancel()
	return nil
}
