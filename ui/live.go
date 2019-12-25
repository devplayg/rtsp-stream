package ui

func LivePage() string {
	return `{{define "content"}}
    <div id="cameras">
        <div class="row row-cols-3">
			{{range .streams }}
            <div class="col">
                	<video-js id="live{{ .Id }}" class="vjs-default-skin vjs-fluid"><source></video-js>
            </div>
			{{end}}
        </div>
    </div>
{{end}}

{{define "script"}}
	<script src="/static/assets/modules/stream/formatter.js"></script>
	<script src="/static/assets/modules/stream/live.js"></script>
	<script>
		let streams = [];
		{{range .streams }}
        streams.push({
			id: {{.Id}},
			enabled: {{.Enabled}},
			recording: {{.Recording}},
			status: {{.Status}},
		});
		{{end}}
	</script>
{{end}}
`
	// w.Header().Set("Content-Type", mime.TypeByExtension(".html"))
}
