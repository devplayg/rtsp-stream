package ui

import (
	"fmt"
	"html/template"
	"net/http"
)

func Stream(w http.ResponseWriter, r *http.Request) {
	tmpl := template.New("stream")

	tmpl, err := tmpl.Parse(Base())
	if err != nil {

	}

	//var err error
	//if tmpl, err = tmpl.Parse(page); err != nil {
	//	fmt.Println(err)
	//}
	//if tmpl, err = tmpl.Parse(tags); err != nil {
	//	fmt.Println(err)
	//}
	//if tmpl, err = tmpl.Parse(comment); err != nil {
	//	fmt.Println(err)
	//}
	if tmpl, err = tmpl.Parse(StreamPage()); err != nil {
		fmt.Println(err)
	}
	tmpl.Execute(w, nil)

	//tmpl.Execute(os.Stdout, pagedata)
}

func StreamPage() string {
	return `
{{define "content"}}
<div id="app">
{{"{{"}}message{{"}}"}}  
</div>
{{end}}

{{define "script"}}
<script>
var app = new Vue({
  el: '#app',
  data: {
    message: 'Hello Vue'
  }
})
</script>
{{end}}
`

}
