package server

import (
	"encoding/json"
	"errors"
	"github.com/boltdb/bolt"
	"github.com/devplayg/rtsp-stream/common"
	"github.com/devplayg/rtsp-stream/streaming"
)

func GetValueFromDB(bucket, key []byte) ([]byte, error) {
	var data []byte
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return errors.New("bucket not found: " + string(bucket))
		}
		data = b.Get(key)
		return nil
	})
	return data, err
}

func PutDataInDB(bucket, key, value []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return errors.New("bucket not found: " + string(bucket))
		}
		return b.Put(key, value)
	})
}

func IssueStreamId() (int64, error) {
	var streamId int64
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(common.StreamBucket)
		id, _ := b.NextSequence()
		streamId = int64(id)
		return b.Put(CreateStreamKey(streamId), nil)
	})
	return streamId, err
}

func CreateStreamKey(id int64) []byte {
	return common.Int64ToBytes(id)
}

func SaveStreamInDB(stream *streaming.Stream) error {
	return db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(streaming.GetStreamBucketName(stream.Id, "")); err != nil {
			return err
		}

		b := tx.Bucket(common.StreamBucket)
		buf, err := json.Marshal(stream)
		if err != nil {
			return err
		}
		return b.Put(CreateStreamKey(stream.Id), buf)
	})
}

func DeleteStreamInDB(id int64) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(common.StreamBucket)
		return b.Delete(CreateStreamKey(id))
	})
}
