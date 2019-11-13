package server

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type Controller struct {
	server *Server
	router *mux.Router
}

func (c *Controller) init() {
	r := mux.NewRouter()
	r.HandleFunc("/streams", c.GetStreams).Methods("GET")
	r.HandleFunc("/streams", c.PostStream).Methods("POST")
	http.Handle("/", r)
	c.router = r
}

func NewController(server *Server) *Controller {
	controller := Controller{
		server: server,
	}
	controller.init()
	return &controller
}

func (c *Controller) GetStreams(w http.ResponseWriter, r *http.Request) {
	list, err := c.server.getAllStreams()
	if err != nil {
		log.Error(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(list)
}

/*
 curl -i -X POST -d '{"url":"rtsp://127.0.0.1:30101/Streaming/Channels/101/","username":"admin","password":"1234"}' http://192.168.0.14:9000/streams
*/

func (c *Controller) PostStream(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Error(err)
		return
	}

	spew.Dump(r.Form.Get("username"))

	stream := NewStream("rtsp")
	c.server.AddStream(stream)
	w.WriteHeader(http.StatusOK)
}
