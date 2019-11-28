# RTSP Streaming

### Powered by 

* Go
* FFmpeg
* BoltDB
* Highwayhash
* Hippo


### Structure

```
+----------------------------------------+  
|            |  stream | stream | stream |          +-----------+
|            |  assist | assist | assist |          |           |
|----------------------------------------|          |   Minio   |
| controller |           manager         |          |           |
|----------------------------------------|          +-----------+
|               server                   |   
+----------------------------------------+
```

Server

- Server framework

Database

- Key/Value database
- BoltDB (https://github.com/boltdb/bolt)

Manager

- Streaming management
- Start, stop, add, and remove streaming

Stream

- Streaming object

Assistant

- Stream's assistant
- Check streaming status
- Archive live videos

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
