package ui

func VideoPage() string {
	return `{{define "content"}}
    <div class="row">
        <div class="col">
            <table  id="table-videos"
                    data-toolbar="#toolbar-videos"
                    data-search="true"
                    data-id-field="id"
                    data-pagination="true"
                    data-side-pagination="client"
                    data-show-refresh="true"
                    data-sort-name="date"
                    data-sort-order="asc">
                <thead>
            </table>
        </div>
        <div class="col">
            <video-js id="player" class="vjs-default-skin vjs-fluid">
            </video-js>
        </div>
    </div>
{{end}}

{{define "script"}}
	<script src="/static/assets/modules/stream/formatter.js"></script>
	<script src="/static/assets/modules/stream/videos.js"></script>
{{end}}
`

}
