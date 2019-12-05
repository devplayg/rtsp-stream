package streaming

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/devplayg/rtsp-stream/common"
	"github.com/devplayg/rtsp-stream/ui"
	"github.com/gorilla/mux"
	"github.com/minio/minio-go"
	log "github.com/sirupsen/logrus"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Controller struct {
	router    *mux.Router
	server    *Server
	manager   *Manager
	staticDir string
}

func (c *Controller) init() {
	r := mux.NewRouter()
	r.HandleFunc("/streams", c.GetStreams).Methods("GET")
	r.HandleFunc("/streams", c.AddStream).Methods("POST")
	r.HandleFunc("/streams/debug", c.DebugStream).Methods("GET")

	r.HandleFunc("/streams/{id:[0-9]+}", c.GetStreamById).Methods("GET")
	r.HandleFunc("/streams/{id:[0-9]+}", c.UpdateStream).Methods("PATCH")
	r.HandleFunc("/streams/{id:[0-9]+}", c.DeleteStream).Methods("DELETE")

	r.HandleFunc("/streams/{id:[0-9]+}/start", c.StartStream).Methods("GET")
	r.HandleFunc("/streams/{id:[0-9]+}/stop", c.StopStream).Methods("GET")

	// Today
	// http://127.0.0.1:8000/videos/1/today/m3u8
	r.HandleFunc("/videos/{id:[0-9]+}/today/m3u8", c.GetTodayM3u8).Methods("GET")
	// http://127.0.0.1:8000/videos/1/today/media0.ts
	r.HandleFunc("/videos/{id:[0-9]+}/today/{media}.ts", c.GetTodayVideo).Methods("GET")

	// Live
	// http://127.0.0.1:8000/videos/1/live/m3u8
	r.HandleFunc("/videos/{id:[0-9]+}/live/m3u8", c.GetLiveM3u8).Methods("GET")
	//  http://127.0.0.1:8000/videos/1/live/media0.ts
	r.HandleFunc("/videos/{id:[0-9]+}/live/{media}.ts", c.GetLiveVideo).Methods("GET")

	r.HandleFunc("/videos/{id:[0-9]+}/date/{date:[0-9]+}/m3u8", c.GetDailyM3u8).Methods("GET")
	r.HandleFunc("/videos/{id:[0-9]+}/date/{date:[0-9]+}/{media}.ts", c.GetDailyVideo).Methods("GET")
	//http://127.0.0.1:8000/videos/1/date/20191126/1.ts

	r.
		PathPrefix("/static").
		Handler(http.StripPrefix("/static", http.FileServer(http.Dir(c.staticDir))))

	r.HandleFunc("/streams/", ui.Stream).Methods("GET")
	//http.HandleFunc("/ui", serveTemplate)
	http.Handle("/", r)

	c.router = r
}

func NewController(server *Server) *Controller {
	controller := Controller{
		server:    server,
		manager:   server.manager,
		staticDir: server.config.Static.Dir,
	}
	controller.init()
	return &controller
}

/*
	curl -i -X POST -d '{"uri":"rtsp://127.0.0.1:30101/Streaming/Channels/101/","username":"admin","password":"xxxx"}' http://192.168.0.14:9000/streams
*/
func (c *Controller) AddStream(w http.ResponseWriter, r *http.Request) {
	stream, err := parseAndGetStream(r.Body)
	if err != nil {
		Response(w, err, http.StatusBadRequest)
		return
	}

	err = c.manager.addStream(stream)
	if err != nil {
		Response(w, err, http.StatusBadRequest)
		return
	}

	Response(w, nil, http.StatusOK)
}

func (c *Controller) GetStreams(w http.ResponseWriter, r *http.Request) {
	list := c.manager.getStreams()
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		Response(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", common.ContentTypeJson)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Write(data)
}

func (c *Controller) UpdateStream(w http.ResponseWriter, r *http.Request) {
	stream, err := parseAndGetStream(r.Body)
	if err != nil {
		Response(w, err, http.StatusBadRequest)
		return
	}

	err = c.manager.updateStream(stream)
	if err != nil {
		Response(w, err, http.StatusBadRequest)
		return
	}

	Response(w, nil, http.StatusOK)
}

/*
	curl -i -X DELETE http://192.168.0.14:9000/streams/ee3b86ddc65b2dcbf7edcc649825af2c
*/
func (c *Controller) DeleteStream(w http.ResponseWriter, r *http.Request) {
	streamId, err := parseAndGetStreamId(r)
	if err != nil {
		Response(w, err, http.StatusBadRequest)
		return
	}

	err = c.manager.deleteStream(streamId)
	if err != nil {
		Response(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c *Controller) StartStream(w http.ResponseWriter, r *http.Request) {
	streamId, err := parseAndGetStreamId(r)
	if err != nil {
		Response(w, err, http.StatusBadRequest)
		return
	}

	err = c.manager.startStreaming(streamId, "request from "+r.RemoteAddr)
	if err != nil {
		Response(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c *Controller) GetStreamById(w http.ResponseWriter, r *http.Request) {
	streamId, err := parseAndGetStreamId(r)
	if err != nil {
		Response(w, err, http.StatusBadRequest)
		return
	}

	stream := c.manager.getStreamById(streamId)
	data, err := json.MarshalIndent(stream, "", "  ")
	if err != nil {
		Response(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", common.ContentTypeJson)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Write(data)
}

func (c *Controller) GetTodayM3u8(w http.ResponseWriter, r *http.Request) {
	streamId, err := parseAndGetStreamId(r)
	if err != nil {
		Response(w, err, http.StatusBadRequest)
		return
	}

	tags, err := c.manager.getM3u8(streamId, "")
	if err != nil {
		Response(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", common.ContentTypeJson)
	w.Header().Set("Content-Length", strconv.Itoa(len(tags)))
	w.Write([]byte(tags))
}

func (c *Controller) GetLiveM3u8(w http.ResponseWriter, r *http.Request) {
	streamId, err := parseAndGetStreamId(r)
	if err != nil {
		Response(w, err, http.StatusBadRequest)
		return
	}

	path := filepath.Join(c.server.config.Storage.LiveDir, strconv.FormatInt(streamId, 10), common.LiveM3u8FileName)
	http.ServeFile(w, r, path)

	//file, err := os.Open(path)
	//if err != nil {
	//	Response(w, err, http.StatusBadRequest)
	//	return
	//}
	//stat, err := file.Stat()
	//if err != nil {
	//	Response(w, err, http.StatusInternalServerError)
	//	return
	//}
	//
	//w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
	//if _, err = io.Copy(w, file); err != nil{
	//	Response(w, err, http.StatusInternalServerError)
	//	return
	//}
}

func (c *Controller) GetLiveVideo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	path := filepath.ToSlash(filepath.Join(c.server.config.Storage.LiveDir, vars["id"], vars["media"]+".ts"))

	//streamId, err := parseAndGetStreamId(r)
	//if err != nil {
	//	Response(w, err, http.StatusBadRequest)
	//	return
	//}
	//
	//path := filepath.Join(c.server.liveDir, strconv.FormatInt(streamId, 10), LiveM3u8FileName)
	http.ServeFile(w, r, path)
}

func (c *Controller) StopStream(w http.ResponseWriter, r *http.Request) {
	streamId, err := parseAndGetStreamId(r)
	if err != nil {
		Response(w, err, http.StatusBadRequest)
		return
	}

	err = c.manager.stopStreaming(streamId)
	if err != nil {
		Response(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", common.ContentTypeJson)
	w.WriteHeader(http.StatusOK)
}

func (c *Controller) DebugStream(w http.ResponseWriter, r *http.Request) {
	_ = common.DB.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(common.StreamBucket))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			log.Debugf("[%s] %s", string(k), string(v))
		}
		return nil
	})
}

func (c *Controller) GetTodayVideo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	path := filepath.ToSlash(filepath.Join(c.server.config.Storage.LiveDir, vars["id"], vars["media"]+".ts"))
	file, err := os.Open(path)
	if err != nil {
		Response(w, err, http.StatusInternalServerError)
		return
	}

	stat, err := file.Stat()
	if err != nil {
		Response(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Accept-Range", "bytes")
	w.Header().Set("Content-Type", common.ContentTypeTs)
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
	if _, err = io.Copy(w, file); err != nil {
		Response(w, err, http.StatusInternalServerError)
		return
	}
}

//func (c *Controller) GetM3u8(w http.ResponseWriter, r *http.Request) {
//	vars := mux.Vars(r)
//	bucketName := VideoRecordBucket
//	objectName := filepath.ToSlash(filepath.Join(vars["id"], vars["date"], VideoFilePrefix+vars["seq"]+".ts"))
//
//	object, err := MinioClient.GetObject(bucketName, objectName, minio.GetObjectOptions{})
//	if err != nil {
//		Response(w, err, http.StatusInternalServerError)
//		return
//	}
//
//	//reader := bufio.NewReader(object)
//	//s, _ := object.Stat()
//	//s.Size
//	//info, _ := object.Stat()
//
//	w.Header().Set("Accept-Range", "bytes")
//	w.Header().Set("Content-Type", "video/vnd.dlna.mpeg-tts")
//
//	//if _, err = io.Copy(w, object); err != nil{
//	//    Response(w, err, http.StatusInternalServerError)
//	//    return
//	//}
//
//	buf := new(bytes.Buffer)
//	n, err := buf.ReadFrom(object)
//	if err != nil {
//		Response(w, err, http.StatusInternalServerError)
//		return
//	}
//	w.Header().Set("Content-Length", strconv.FormatInt(n, 10))
//	w.WriteHeader(http.StatusOK)
//	w.Write(buf.Bytes())
//
//	//Accept-Ranges: bytes
//	//Content-Length: 1099988
//	//Content-Type: video/vnd.dlna.mpeg-tts
//	//Date: Tue, 26 Nov 2019 10:50:15 GMT
//	//Last-Modified: Tue, 26 Nov 2019 10:34:51 GMT
//
//	//w.WriteHeader(http.StatusOK)
//	//b := bytes.NewBuffer(object)
//	//bufre
//
//	//fmt.Fprintf()
//	//localFile, err := os.Create("/tmp/local-file.jpg")
//	//if err != nil {
//	//    fmt.Println(err)
//	//    return
//	//}
//	//if _, err = io.Copy(localFile, object); err != nil {
//	//   fmt.Println(err)
//	//   return
//	//}
//
//}

func (c *Controller) GetDailyVideo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	objectName := filepath.ToSlash(filepath.Join(vars["id"], vars["date"], vars["media"]+".ts"))
	object, err := common.MinioClient.GetObject(common.VideoRecordBucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		Response(w, err, http.StatusInternalServerError)
		return
	}

	//reader := bufio.NewReader(object)
	//s, _ := object.Stat()
	//s.Size
	//info, _ := object.Stat()

	w.Header().Set("Accept-Range", "bytes")
	w.Header().Set("Content-Type", common.ContentTypeTs)

	//if _, err = io.Copy(w, object); err != nil{
	//    Response(w, err, http.StatusInternalServerError)
	//    return
	//}

	buf := new(bytes.Buffer)
	n, err := buf.ReadFrom(object)
	if err != nil {
		Response(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Length", strconv.FormatInt(n, 10))
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())

	//Accept-Ranges: bytes
	//Content-Length: 1099988
	//Content-Type: video/vnd.dlna.mpeg-tts
	//Date: Tue, 26 Nov 2019 10:50:15 GMT
	//Last-Modified: Tue, 26 Nov 2019 10:34:51 GMT

	//w.WriteHeader(http.StatusOK)
	//b := bytes.NewBuffer(object)
	//bufre

	//fmt.Fprintf()
	//localFile, err := os.Create("/tmp/local-file.jpg")
	//if err != nil {
	//    fmt.Println(err)
	//    return
	//}
	//if _, err = io.Copy(localFile, object); err != nil {
	//   fmt.Println(err)
	//   return
	//}

}

//func (c *Controller) RedirectToVideoFile(w http.ResponseWriter, r *http.Request) {
//	vars := mux.Vars(r)
//	seq, _ := strconv.ParseInt(vars["seq"], 10, 16)
//	var data []byte
//	//streamId, _ := strconv.ParseInt(vars["id"], 10, 16)
//	//date := vars["date"]
//	bucket := []byte(fmt.Sprintf("stream-%s-%s", vars["id"], vars["date"]))
//
//	err := DB.View(func(tx *bolt.Tx) error {
//		b := tx.Bucket(bucket)
//		if b == nil {
//			return nil
//		}
//
//		data = b.Get(Int64ToBytes(seq))
//
//		//spew.Dump(data)
//
//		//c := b.Cursor()
//		//
//		//for k, v := c.First(); k != nil; k, v = c.Next() {
//		//   var videoRecord VideoRecord
//		//   err := json.Unmarshal(v, &videoRecord)
//		//   if err != nil {
//		//       log.Error(err)
//		//       continue
//		//   }
//
//		//    if videoRecord.Duration > maxTargetDuration {
//		//        maxTargetDuration = videoRecord.Duration
//		//    }
//		//    if firstSeq < 1 {
//		//        firstSeq = BytesToInt64(k)
//		//    }
//		//
//		//    body += fmt.Sprintf("#EXTINF:%.6f,\n", videoRecord.Duration)
//		//    body += fmt.Sprintf("%d.ts\n", BytesToInt64(k))
//
//		//keys = append(keys, BytesToInt64(k))
//		//videos = append(videos, &videoRecord)
//		//}
//		return nil
//	})
//
//	if data == nil {
//		Response(w, errors.New("no data"), http.StatusBadRequest)
//		return
//	}
//	spew.Dump(data)
//	var videoRecord VideoRecord
//	err = json.Unmarshal(data, &videoRecord)
//	if err != nil {
//		Response(w, err, http.StatusInternalServerError)
//		return
//	}
//
//	if err != nil {
//		Response(w, err, http.StatusInternalServerError)
//		return
//	}
//
//	if len(videoRecord.Url) < 1 {
//		Response(w, errors.New("no data"), http.StatusBadRequest)
//		return
//	}
//
//	http.Redirect(w, r, videoRecord.Url, http.StatusSeeOther)
//
//}

func (c *Controller) GetDailyM3u8(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	objectName := fmt.Sprintf("%s/%s/%s", vars["id"], vars["date"], "xxxx")
	object, err := common.MinioClient.GetObject(common.VideoRecordBucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"bucket": common.VideoRecordBucket,
			"object": objectName,
		}).Debugf("failed to get object from Minio")
		Response(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Accept-Range", "bytes")
	//w.Header().Set("Content-Type", "video/vnd.dlna.mpeg-tts")
	//w.Header().Set("Content-Type", ContentTypeM3u8)

	//if _, err = io.Copy(w, object); err != nil{
	//    Response(w, err, http.StatusInternalServerError)
	//    return
	//}
	buf := new(bytes.Buffer)
	n, err := buf.ReadFrom(object)
	if err != nil {
		Response(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Length", strconv.FormatInt(n, 10))
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

//func (c *Controller) GetM3u8(w http.ResponseWriter, r *http.Request) {
//	vars := mux.Vars(r)
//	streamId, _ := strconv.ParseInt(vars["id"], 10, 16)
//	date := vars["date"]
//	bucket := []byte(fmt.Sprintf("stream-%d-%s", streamId, date))
//	var maxTargetDuration float32
//	var firstSeq int64
//	//videos := make([]*VideoRecord, 0)
//	//keys := make([]int64, 0)
//
//	body := ""
//	err := DB.View(func(tx *bolt.Tx) error {
//		// Assume bucket exists and has keys
//		b := tx.Bucket(bucket)
//		if b == nil {
//			return nil
//		}
//
//		c := b.Cursor()
//
//		for k, v := c.First(); k != nil; k, v = c.Next() {
//			var videoRecord VideoRecord
//			err := json.Unmarshal(v, &videoRecord)
//			if err != nil {
//				log.Error(err)
//				continue
//			}
//
//			if videoRecord.Duration > maxTargetDuration {
//				maxTargetDuration = videoRecord.Duration
//			}
//			if firstSeq < 1 {
//				firstSeq = BytesToInt64(k)
//			}
//
//			body += fmt.Sprintf("#EXTINF:%.6f,\n", videoRecord.Duration)
//			body += fmt.Sprintf("media%d.ts\n", BytesToInt64(k))
//
//			//keys = append(keys, BytesToInt64(k))
//			//videos = append(videos, &videoRecord)
//		}
//		return nil
//	})
//	if err != nil {
//		Response(w, err, http.StatusInternalServerError)
//		return
//	}
//	//sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
//	//sort.SliceStable(videos, func(i, j int) bool {
//	//   return videos[i].UnixTime < videos[j].UnixTime
//	//})
//
//	//for _, v := range videos {
//	//if v.Duration > maxTargetDuration {
//	//    maxTargetDuration = v.Duration
//	//}
//	//    body += fmt.Sprintf("#EXTINF:%.6f,\n", v.Duration)
//	//    body += v.Url+ "\n"
//	//}
//	m3u8 := GetM3u8Header(firstSeq, math.Ceil(float64(maxTargetDuration))) + body + GetM3u8Footer()
//	//w.Header().Set("Access-Control-Allow-Origin", "*")
//	//w.Header().Set("Access-Control-Allow-Methods", "GET")
//	//w.Header().Set("Cache-Control", "no-cache")
//	w.Header().Set("Content-Length", strconv.Itoa(len(m3u8)))
//	//w.Header().Set("Accept-Ranges", "bytes")
//	//w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
//	w.WriteHeader(http.StatusOK)
//
//	// application/vnd.apple.mpegurl
//
//	//url, _ := url.Parse(url.QueryEscape(str))
//	//if err != nil { panic(err) }
//	//fmt.Println(url.String())
//
//	//fmt.Fprintf(w, m3u8)
//	if _, err = w.Write([]byte(m3u8)); err != nil {
//		log.Error(err)
//	}
//}

//
//func Download(file string, filename ...string) {
//    // check get file error, file not found or other error.
//    if _, err := os.Stat(file); err != nil {
//        http.ServeFile(output.Context.ResponseWriter, output.Context.Request, file)
//        return
//    }
//
//    var fName string
//    if len(filename) > 0 && filename[0] != "" {
//        fName = filename[0]
//    } else {
//        fName = filepath.Base(file)
//    }
//    //https://tools.ietf.org/html/rfc6266#section-4.3
//    fn := url.PathEscape(fName)
//    if fName == fn {
//        fn = "filename=" + fn
//    } else {
//        /**
//          The parameters "filename" and "filename*" differ only in that
//          "filename*" uses the encoding defined in [RFC5987], allowing the use
//          of characters not present in the ISO-8859-1 character set
//          ([ISO-8859-1]).
//        */
//        fn = "filename=" + fName + "; filename*=utf-8''" + fn
//    }
//    output.Header("Content-Disposition", "attachment; "+fn)
//    output.Header("Content-Description", "File Transfer")
//    output.Header("Content-Type", "application/octet-stream")
//    output.Header("Content-Transfer-Encoding", "binary")
//    output.Header("Expires", "0")
//    output.Header("Cache-Control", "must-revalidate")
//    output.Header("Pragma", "public")
//    http.ServeFile(output.Context.ResponseWriter, output.Context.Request, file)
//}

//

func formatAsDollars(valueInCents int) (string, error) {
	dollars := valueInCents / 100
	cents := valueInCents % 100
	return fmt.Sprintf("$%d.%2d", dollars, cents), nil
}

func formatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d/%d/%d", day, month, year)
}

func urgentNote(acc ui.Account) string {
	return fmt.Sprintf("You have earned 100 VIP points that can be used for purchases")
}

func serveTemplate2(w http.ResponseWriter, r *http.Request) {
	fmap := template.FuncMap{
		"formatAsDollars": formatAsDollars,
		"formatAsDate":    formatAsDate,
		"urgentNote":      urgentNote,
	}

	// Create a new template and parse the letter into it.
	str := "hello"

	//t := template.Must(template.New("email.tmpl").Funcs(fmap).Parse(ui.Layout(str)))
	t := template.Must(template.New("streams").Funcs(fmap).Parse(ui.Layout(str)))
	err := t.Execute(w, ui.CreateMockStatement())
	if err != nil {
		log.Println("executing template:", err)
	}
}
