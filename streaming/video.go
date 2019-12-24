package streaming

import (
	"fmt"
	"github.com/devplayg/rtsp-stream/ui"
	"html/template"
	"mime"
	"net/http"
)

func DisplayVideos(w http.ResponseWriter, r *http.Request) {
	tmpl := template.New("videos")

	tmpl, err := tmpl.Parse(ui.Base(ui.Fluid))
	if err != nil {

	}
	if tmpl, err = tmpl.Parse(ui.Videos()); err != nil {
		fmt.Println(err)
	}
	w.Header().Set("Content-Type", mime.TypeByExtension(".html"))
	tmpl.Execute(w, nil)
}
