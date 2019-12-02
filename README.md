# RTSP Streaming V2

### Powered by

* Go
* FFmpeg
* BoltDB
* MinIO
* Highwayhash
* Hippo


### Structure

```
+-------------------------------------------------+
|                     |  stream | stream | stream |          +-----------+
|                     |  assist | assist | assist |          |           |
|-------------------------------------------------|          |   Minio   |
|  db  |  controller  |           manager         |          |           |
|-------------------------------------------------|          +-----------+
|               server                            |
+-------------------------------------------------+
```

### Server

framework

### Database

- Key/Value database
- BoltDB (https://github.com/boltdb/bolt)

### Manager

- manages all streams
- starts, stops, adds, and removes streams

### Stream

receives live stream

### Assistant (Stream's assistant)

helps stream. He is like a slave.

- checks streaming status
- archives live videos and send it to object storage

### Structure

* server
    - boltdb: database
    - controller : handle API
    - manager : streaming manager
        - stream: streaming object
            -  assistant: checking status, merging video files


### Database

|Bucket|Key|Value|
|---|---|---|
|streams|Stream ID (int64)|Stream information (Stream)|
|stream-{id}-{YYYYMMDD}|media file name (string)|Media information (Media)|
|transmission|Stream ID|TransmissionResult|
