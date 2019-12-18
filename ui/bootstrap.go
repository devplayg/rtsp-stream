package ui

func BootstrapBase() string {
	return `<!doctype html>
<html lang="en">
  <head>
    <!-- Required meta tags -->
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">

    <!-- Bootstrap CSS -->
    <link rel="stylesheet" href="/assets/css/bootstrap.min.css">
    <title>Hello, world!</title>
  </head>
  <body>
    {{ block "content" . }}{{ end }}
    <script src="/assets/js/jquery-3.4.1.min.js" ></script>
    <script src="/assets/js/popper.min.js"></script>
    <script src="/assets/js/bootstrap.min.js" ></script>
    {{ block "script" . }}{{ end }}
  </body>
</html>`
}

func Videos() string {
	return `
{{define "content"}}
<button class="btn btn-default">TEST</button>
<div class="row">
    <div class="col-lg-3">
        col1
    </div>
<div class="col-lg-3">
        col1
    </div>
<div class="col-lg-3">
        col1
    </div>
<div class="col-lg-3">
        col1
    </div>
</div>
{{end}}

{{define "script"}}
<script>
console.log(3233);
</script>
{{end}}
`

}
