package server

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

type Controller struct {
	server  *Server
	manager *Manager

	router *mux.Router
}

func (c *Controller) init() {
	r := mux.NewRouter()
	r.HandleFunc("/streams", c.GetStreams).Methods("GET")
	r.HandleFunc("/streams", c.AddStream).Methods("POST")
	r.HandleFunc("/streams/{id:[0-9]+}", c.DeleteStream).Methods("DELETE")
	http.Handle("/", r)
	c.router = r
}

func NewController(server *Server, manager *Manager) *Controller {
	controller := Controller{
		server:  server,
		manager: manager,
	}
	controller.init()
	return &controller
}

func (c *Controller) GetStreams(w http.ResponseWriter, r *http.Request) {
	list, err := c.manager.getAllStreams()
	if err != nil {
		ResponseError(w, err, http.StatusOK)
		return
	}

	json, err := json.Marshal(list)
	if err != nil {
		ResponseError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", ApplicationJson)
	w.Write(json)
}

/*
	curl -i -X POST -d '{"uri":"rtsp://58.72.99.132:30101/Streaming/Channels/101/","username":"admin","password":"unisem1234"}' http://192.168.0.32:9000/streams
*/
func (c *Controller) AddStream(w http.ResponseWriter, r *http.Request) {
	stream := &Stream{}
	err := c.checkStreamRequest(r.Body, stream)
	if err != nil {
		ResponseError(w, err, http.StatusBadRequest)
		return
	}
	err = c.manager.AddStream(stream)
	if err != nil {
		ResponseError(w, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

/*
	curl -i -X DELETE http://192.168.0.32:9000/streams/1
*/
func (c *Controller) DeleteStream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		ResponseError(w, err, http.StatusBadRequest)
		return
	}
	err = c.manager.DeleteStream(id)
	if err != nil {
		ResponseError(w, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *Controller) checkStreamRequest(body io.Reader, stream *Stream) error {
	uri, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(uri, stream); err != nil {
		return err
	}

	if _, err := url.Parse(stream.Uri); err != nil {
		return ErrorInvalidUri
	}

	return nil
}
