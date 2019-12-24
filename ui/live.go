package ui

func LivePage() string {
	return `{{define "content"}}
    <div id="cameras">
        <div class="row row-cols-3">
            <div class="col">
                <video-js id="live1" class="vjs-default-skin vjs-fluid"><source></video-js>
            </div>
        </div>
    </div>
{{end}}

{{define "script"}}
	<script src="/static/assets/modules/stream/formatter.js"></script>
	<script src="/static/assets/modules/stream/live.js"></script>
{{end}}
`

}
