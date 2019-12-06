package streaming

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/devplayg/rtsp-stream/common"
)

func GetStream(id int64) (*Stream, error) {
	var stream *Stream
	err := common.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(common.StreamBucket)
		data := b.Get(common.Int64ToBytes(id))
		if data == nil {
			return nil
		}

		err := json.Unmarshal(data, &stream)
		if err != nil {
			return err
		}

		return nil
	})

	return stream, err
}

func IssueStreamId() (int64, error) {
	var streamId int64
	err := common.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(common.StreamBucket)
		id, _ := b.NextSequence()
		streamId = int64(id)
		return b.Put(common.Int64ToBytes(streamId), nil)
	})
	return streamId, err
}

func SaveStreamInDB(stream *Stream) error {
	return common.DB.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(GetStreamBucketName(stream.Id, "")); err != nil {
			return err
		}

		b := tx.Bucket(common.StreamBucket)
		buf, err := json.Marshal(stream)
		if err != nil {
			return err
		}
		return b.Put(common.Int64ToBytes(stream.Id), buf)
	})
}

func DeleteStreamInDB(id int64) error {
	return common.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(common.StreamBucket)
		return b.Delete(common.Int64ToBytes(id))
	})
}
