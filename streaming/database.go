package streaming

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/devplayg/rtsp-stream/utils"
)

func GetStream(id int64) (*Stream, error) {
	var stream *Stream
	err := DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(StreamBucket)
		data := b.Get(utils.Int64ToBytes(id))
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

func SaveStream(stream *Stream) error {
	return DB.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(GetStreamBucketName(stream.Id, "")); err != nil {
			return err
		}

		b := tx.Bucket(StreamBucket)
		buf, err := json.Marshal(stream)
		if err != nil {
			return err
		}
		return b.Put(utils.Int64ToBytes(stream.Id), buf)
	})
}

func DeleteStream(id int64) error {
	return DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(StreamBucket)
		return b.Delete(utils.Int64ToBytes(id))
	})
}
