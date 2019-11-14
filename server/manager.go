package server

import (
    "encoding/binary"
    "encoding/json"
    "github.com/boltdb/bolt"
    log "github.com/sirupsen/logrus"
    "sync"
)

type Manager struct {
    server    *Server
    streamMap sync.Map
}

func NewManager(server *Server) *Manager {
    return &Manager{
        server: server,
    }
}

func (m *Manager) getAllStreams() ([]*Stream, error) {

    streams := make([]*Stream, 0)

    err := m.server.db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket(StreamBucket)
        c := b.Cursor()
        for k, v := c.First(); k != nil; k, v = c.Next() {
            var stream Stream
            err := json.Unmarshal(v, &stream)
            if err != nil {
                log.Error(err)
                continue
            }
            streams = append(streams, &stream)
        }

        return nil
    })

    return streams, err

    //streams := make([]*Stream, 0)
    //m.streamMap.Range(func(k, v interface{}) bool {
    //	streams = append(streams, v.(*Stream))
    //	//fmt.Printf("key: %s, value: %s\n", k, v) // key: hoge, value: fuga
    //	return true
    //})
    //
    //return json.Marshal(streams)
}

func (m *Manager) AddStream(stream *Stream) error {
    return m.SaveStream(stream)
}

func (m *Manager) DeleteStream(id int64) error {
    return m.server.db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(StreamBucket))
        return b.Delete(Int64ToBytes(id))
    })
}


//func (m *Manager) SaveStatus() error {
//}

func (m *Manager) SaveStream(stream *Stream) error {
    return m.server.db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(StreamBucket))
        id, _ := b.NextSequence()
        stream.Id = int64(id)

        // Marshal user data into bytes.
        buf, err := json.Marshal(stream)
        if err != nil {
            return err
        }

        // Persist bytes to users bucket.
        return b.Put(Int64ToBytes(stream.Id), buf)
    })
}

func BytesToInt64(buf []byte) int64 {
    return int64(binary.BigEndian.Uint64(buf))
}

func Int64ToBytes(i int64) []byte {
    var buf = make([]byte, 8)
    binary.BigEndian.PutUint64(buf, uint64(i))
    return buf
}
