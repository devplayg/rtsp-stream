# RTSP Streaming V2

### Powered by

* Go
* FFmpeg
* BoltDB
* MinIO
* Highwayhash
* Hippo(https://github.com/devplayg/hippo)


### Do do

- [X] Write daily video size 
- [X] If stream uri, name or password are changed, reload the stream  




### Structure

```
+-------------------------------------------------+
|                     |  stream | stream | stream |          +----------------+
|                     |  assist | assist | assist |          |                |
|-------------------------------------------------|          |     Minio      |
|  db  |  controller  |           manager         |          | 127.0.0.1:9000 |
|-------------------------------------------------|          |                |
|               server                            |          +----------------+
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
- watches all streams

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

server.db

|Bucket|Key|Value|
|---|---|---|
|streams|Stream ID (int64)|Stream information (Stream)|
|video-{id}|YYYYMMDD|Video|
|config|string|string|

stream-{id}.db
|Bucket|Key|Value|
|---|---|---|
|{YYYYMMDD}|media file name (string)|Media information (Media)|