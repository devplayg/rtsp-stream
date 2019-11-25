package ui

func Layout() string {
	return `{{define "layout"}}
    <!doctype html>
    <html>
    <head>
        <meta charset="utf-8">
        <title>{{template "title"}}</title>
    </head>
    <body>
    {{template "body"}}

    </body>
    </html>
{{end}}`
}

func Hello() string {
	return `{{define "title"}}A templated page{{end}}

{{define "body"}}
    <h1>Hello from a templated page</h1>
{{end}}`
}
