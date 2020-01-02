package server

import (
	"errors"
	"github.com/boltdb/bolt"
	"github.com/devplayg/rtsp-stream/common"
)

func GetValueFromDbBucket(bucket, key []byte) ([]byte, error) {
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

func PutDataIntoDbBucket(bucket, key, value []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return errors.New("bucket not found: " + string(bucket))
		}
		return b.Put(key, value)
	})
}

func GetDbBuckets(db *bolt.DB) ([][]byte, error) {
	buckets := make([][]byte, 0)
	err := db.View(func(tx *bolt.Tx) error {
		err := tx.ForEach(func(b []byte, _ *bolt.Bucket) error {
			buckets = append(buckets, b)
			return nil
		})
		return err
	})
	return buckets, err
}

func IssueStreamId() (int64, error) {
	var streamId int64
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(common.StreamBucket)
		id, _ := b.NextSequence()
		streamId = int64(id)
		return b.Put(common.CreateStreamKey(streamId), nil)
	})
	return streamId, err
}

//
//func SaveStreamInDB(stream *streaming.Stream) error {
//	return db.Update(func(tx *bolt.Tx) error {
//		if _, err := tx.CreateBucketIfNotExists(streaming.GetStreamBucketName(stream.Id, "")); err != nil {
//			return err
//		}
//
//		b := tx.Bucket(common.StreamBucket)
//		buf, err := json.Marshal(stream)
//		if err != nil {
//			return err
//		}
//		return b.Put(CreateStreamKey(stream.Id), buf)
//	})
//}

//func SaveStreamInDB(stream *streaming.Stream) error {
//	return db.Update(func(tx *bolt.Tx) error {
//		if _, err := tx.CreateBucketIfNotExists(streaming.GetStreamBucketName(stream.Id, "")); err != nil {
//			return err
//		}
//
//		b := tx.Bucket(common.StreamBucket)
//		buf, err := json.Marshal(stream)
//		if err != nil {
//			return err
//		}
//		return b.Put(CreateStreamKey(stream.Id), buf)
//	})
//}

func DeleteStreamInDB(id int64) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(common.StreamBucket)
		return b.Delete(common.CreateStreamKey(id))
	})
}
