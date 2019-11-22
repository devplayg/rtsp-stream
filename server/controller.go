package server

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var StaticDir = "/static/"

type Controller struct {
	server  *Server
	manager *Manager

	router *mux.Router
}

func (c *Controller) init() {
	r := mux.NewRouter()
	r.HandleFunc("/streams", c.GetStreams).Methods("GET")
	r.HandleFunc("/streams", c.AddStream).Methods("POST")
	r.HandleFunc("/streams/debug", c.DebugStream).Methods("GET")

	r.HandleFunc("/streams/{id}", c.GetStreamById).Methods("GET")
	r.HandleFunc("/streams/{id}", c.UpdateStream).Methods("PATCH")
	r.HandleFunc("/streams/{id}", c.DeleteStream).Methods("DELETE")

	r.HandleFunc("/streams/{id}/start", c.StartStream).Methods("GET")
	r.HandleFunc("/streams/{id}/stop", c.StopStream).Methods("GET")

	r.
		PathPrefix(StaticDir).
		Handler(http.StripPrefix(StaticDir, http.FileServer(http.Dir("."+StaticDir))))

	//http.HandleFunc("/", serveTemplate)

	http.Handle("/", r)
	c.router = r
}

func NewController(server *Server) *Controller {
	controller := Controller{
		server:  server,
		manager: server.manager,
	}
	controller.init()
	return &controller
}

func (c *Controller) GetStreams(w http.ResponseWriter, r *http.Request) {
	list := c.manager.getStreams()
	data, err := json.Marshal(list)
	if err != nil {
		Response(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", ApplicationJson)
	w.Write(data)
}

func (c *Controller) GetStreamById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if len(vars["id"]) < 1 {
		Response(w, errors.New("empty stream key"), http.StatusBadRequest)
		return
	}
	id, _ := strconv.ParseInt(vars["id"], 10, 16)
	stream := c.manager.getStreamById(id)
	data, err := json.Marshal(stream)
	if err != nil {
		Response(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", ApplicationJson)
	w.Write(data)
}

/*
	curl -i -X POST -d '{"uri":"rtsp://127.0.0.1:30101/Streaming/Channels/101/","username":"admin","password":"xxxx"}' http://192.168.0.14:9000/streams
*/
func (c *Controller) AddStream(w http.ResponseWriter, r *http.Request) {
	//log.WithFields(log.Fields{
	//    "ip": r.RemoteAddr,
	//    "uri": r.RequestURI,
	//}).Debug("add stream")
	stream := &Stream{}
	err := c.parseStreamRequest(r.Body, stream)
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

func (c *Controller) UpdateStream(w http.ResponseWriter, r *http.Request) {
	stream := &Stream{}
	err := c.parseStreamRequest(r.Body, stream)
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

func (c *Controller) parseStreamRequest(body io.Reader, stream *Stream) error {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(data, stream); err != nil {
		return err
	}

	stream.Uri = strings.TrimSpace(stream.Uri)

	if _, err := url.Parse(stream.Uri); err != nil {
		return ErrorInvalidUri
	}

	return nil
}

/*
	curl -i -X DELETE http://192.168.0.14:9000/streams/ee3b86ddc65b2dcbf7edcc649825af2c
*/
func (c *Controller) DeleteStream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if len(vars["id"]) < 1 {
		Response(w, errors.New("empty stream key"), http.StatusBadRequest)
		return
	}
	id, _ := strconv.ParseInt(vars["id"], 10, 16)
	err := c.manager.deleteStream(id)
	if err != nil {
		Response(w, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// For test
func (c *Controller) StartStream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if len(vars["id"]) < 1 {
		Response(w, errors.New("empty stream key"), http.StatusBadRequest)
		return
	}
	id, _ := strconv.ParseInt(vars["id"], 10, 16)
	stream := c.manager.getStreamById(id)
	if stream == nil {
		Response(w, errors.New("stream not found"), http.StatusOK)
		return
	}

	err := c.manager.startStreaming(stream)
	if err != nil {
		Response(w, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *Controller) StopStream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if len(vars["id"]) < 1 {
		Response(w, errors.New("empty stream key"), http.StatusBadRequest)
		return
	}
	id, _ := strconv.ParseInt(vars["id"], 10, 16)
	err := c.manager.stopStreaming(id)
	if err != nil {
		Response(w, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *Controller) DebugStream(w http.ResponseWriter, r *http.Request) {
	_ = c.server.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(StreamBucket))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			log.Debugf("[%s] %s", string(k), string(v))
		}

		return nil
	})
}
