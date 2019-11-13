package server

import (
	"encoding/json"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
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
	list, err := c.server.manager.getAllStreams()
	if err != nil {
		log.Error(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(list)
}

/*
 curl -i -X POST -d '{"url":"rtsp://127.0.0.1:30101/Streaming/Channels/101/","username":"admin","password":"1234"}' http://192.168.0.14:9000/streams
*/

func (c *Controller) ResponseError(w http.ResponseWriter, err error, status int) {
	log.Error(err)
	w.Header().Add("Content-Type", ApplicationJson)
	b, _ := json.Marshal(Result{Error: err})
	w.WriteHeader(status)
	w.Write(b)
}

func (c *Controller) PostStream(w http.ResponseWriter, r *http.Request) {

	stream := Stream{}
	err := c.checkStreamRequest(r.Body, &stream)
	if err != nil {
		c.ResponseError(w, err, http.StatusBadRequest)
	}

	//if err := r.ParseForm(); err != nil {
	//	log.Error(err)
	//	return
	//}
	//if !c.isAuthenticated(r) {
	//	w.WriteHeader(http.StatusForbidden)
	//	return
	//}
	//var dto StreamDto
	//if err := c.marshalValidatedURI(&dto, r.Body); err != nil {
	//	logrus.Error(err)
	//	c.SendError(w, err, http.StatusBadRequest)
	//	return
	//}

	//spew.Dump(r.Form.Get("username"))

	//stream := NewStream("rtsp")
	c.server.manager.AddStream(stream)
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

	if _, err := url.Parse(stream.URI); err != nil {
		return InvalidUriError
	}
	return nil
}
