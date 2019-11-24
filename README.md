# RTSP Streaming

### Powered by 

* Go
* FFmpeg
* Highwayhash
* BoltDB


### Modules

* server
    - boltdb: database
    - controller : handle API
    - manager : streaming manager
        - stream: streaming object
            -  assistant: checking status, merging video files

### Database

|Bucket|Key|Value|
|---|---|---|
|streams|Stream ID (int64)|Stream object|
|stream-{id}|YYYYMMDD (string)|m3u8 Info.|
|stream-{id}-{YYYYMMDD}|VideoFile (struct)|[]byte{}|
