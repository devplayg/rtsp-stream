package streaming

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
)

type Controller struct {
	server    *Server
	manager   *Manager
	staticDir string
	router    *mux.Router
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
		PathPrefix("/static").
		Handler(http.StripPrefix("/static", http.FileServer(http.Dir(c.staticDir))))

	r.HandleFunc("/streams/", serveTemplate2).Methods("GET")
	//http.HandleFunc("/ui", serveTemplate)
	http.Handle("/", r)

	c.router = r
}

func serveTemplate(w http.ResponseWriter, r *http.Request) {
	lp := filepath.Join("templates", "layout.html")
	fp := filepath.Join("templates", "example.html")

	tmpl, _ := template.ParseFiles(lp, fp)
	tmpl.ExecuteTemplate(w, "layout", nil)
}

func serveTemplate2(w http.ResponseWriter, r *http.Request) {
	//tpl, _ := template.New("layout").Parse(ui.Hello())
	//template.New().
	//tpl.ExecuteTemplate(w, "layout", nil)
	const letter = `
Dear {{.Name}},
{{if .Attended}}
It was a pleasure to see you at the wedding.
{{- else}}
It is a shame you couldn't make it to the wedding.
{{- end}}

{{with .Gift -}}
    Thank you for the lovely {{.}}.
{{end}}
Best wishes,
Josie
---
`

	// Prepare some data to insert into the template.
	//type Recipient struct {
	//    Name, Gift string
	//    Attended   bool
	//}
	//var recipients = []Recipient{
	//    {"Aunt Mildred", "bone china tea set", true},
	//    {"Uncle John", "moleskin pants", false},
	//    {"Cousin Rodney", "", false},
	//}
	//
	//// Create a new template and parse the letter into it.
	//t := template.Must(template.New("l111etter").Parse(ui.Layout()))
	//t.
	//
	//// Execute the template for each recipient.
	//for _, r := range recipients {
	//    err := t.Execute(w, r)
	//    if err != nil {
	//        log.Println("executing template:", err)
	//    }
	//}
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
	stream, err := c.parseStreamRequest(r.Body)
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
	stream, err := c.parseStreamRequest(r.Body)
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

func (c *Controller) parseStreamRequest(body io.Reader) (*Stream, error) {
	stream := NewStream()
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(data, stream); err != nil {
		return nil, err
	}

	stream.Uri = strings.TrimSpace(stream.Uri)

	if _, err := url.Parse(stream.Uri); err != nil {
		return nil, ErrorInvalidUri
	}

	return stream, nil
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
	_ = DB.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(StreamBucket))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			log.Debugf("[%s] %s", string(k), string(v))
		}

		return nil
	})
}

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
