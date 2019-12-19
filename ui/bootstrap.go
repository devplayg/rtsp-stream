package ui

/*
	/assets/css/custom.css
	/assets/img/logo.png
	/assets/js/custom.js
	/assets/js/jquery-3.4.1.min.js
	/assets/js/jquery.mask.min.js
	/assets/js/js.cookie-2.2.1.min.js
	/assets/js/popper.min.js
	/assets/plugins/bootstrap-table/bootstrap-table.min.css
	/assets/plugins/bootstrap-table/bootstrap-table.min.js
	/assets/plugins/bootstrap/bootstrap.min.css
	/assets/plugins/bootstrap/bootstrap.min.js
	/assets/plugins/fontawesome/css/all.min.css
	/assets/plugins/fontawesome/js/all.min.js
	/assets/plugins/fontawesome/webfonts/fa-brands-400.eot
	/assets/plugins/fontawesome/webfonts/fa-brands-400.svg
	/assets/plugins/fontawesome/webfonts/fa-brands-400.ttf
	/assets/plugins/fontawesome/webfonts/fa-brands-400.woff
	/assets/plugins/fontawesome/webfonts/fa-brands-400.woff2
	/assets/plugins/fontawesome/webfonts/fa-regular-400.eot
	/assets/plugins/fontawesome/webfonts/fa-regular-400.svg
	/assets/plugins/fontawesome/webfonts/fa-regular-400.ttf
	/assets/plugins/fontawesome/webfonts/fa-regular-400.woff
	/assets/plugins/fontawesome/webfonts/fa-regular-400.woff2
	/assets/plugins/fontawesome/webfonts/fa-solid-900.eot
	/assets/plugins/fontawesome/webfonts/fa-solid-900.svg
	/assets/plugins/fontawesome/webfonts/fa-solid-900.ttf
	/assets/plugins/fontawesome/webfonts/fa-solid-900.woff
	/assets/plugins/fontawesome/webfonts/fa-solid-900.woff2
	/assets/plugins/moment/moment-timezone-with-data.min.js
	/assets/plugins/moment/moment-timezone.min.js
	/assets/plugins/moment/moment.min.js
	/assets/plugins/videojs/video-js.min.css
	/assets/plugins/videojs/video.min.js
	/assets/plugins/videojs/videojs-http-streaming.min.js
*/

func BootstrapBase() string {
	return `<!doctype html>
<html lang="en">
  <head>
    <!-- Required meta tags -->
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">

    <!-- Bootstrap CSS -->
    <link rel="stylesheet" href="/assets/plugins/videojs/video-js.min.css">
<link rel="stylesheet" href="/assets/plugins/bootstrap/bootstrap.min.css">
    <link rel="stylesheet" href="/assets/css/custom.css">
    <link hrEf="/assets/plugins/fontawesome/css/all.min.css" rel="stylesheet">


    <title>Hello, world!</title>
  </head>
  <body>
    {{ block "content" . }}{{ end }}
    <script src="/assets/js/jquery-3.4.1.min.js" ></script>
    <script src="/assets/js/popper.min.js"></script>
    <script src="/assets/plugins/bootstrap/bootstrap.min.js" ></script>
	<script src="/assets/js/jquery.mask.min.js"></script>
	<script src="/assets/js/js.cookie-2.2.1.min.js"></script>
	<script src="/assets/plugins/bootstrap-table/bootstrap-table.min.css"></script>
	<script src="/assets/plugins/moment/moment-timezone.min.js"></script>
	<script src="/assets/plugins/moment/moment.min.js"></script>
	<script src="/assets/plugins/videojs/video.min.js"></script>
	<script src="/assets/plugins/videojs/videojs-http-streaming.min.js"></script>
	<script src="/assets/js/custom.js"></script>
    {{ block "script" . }}{{ end }}
  </body>
</html>`
}

func Videos() string {
	return `
{{define "content"}}
<button class="btn btn-primary">TEST</button>
<ol>
	<li><i class="fas fa-camera"></i></li>
	<li><i class="fas fa-th"></i></li>
	<li><i class="fas fa-clock"></i></li>
	<li><i class="far fa-clock"></i></li>
</ol>
{{end}}

{{define "script"}}
<script>
console.log(3233);
</script>
{{end}}
`

}
