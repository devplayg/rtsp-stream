package ui

func StreamPage() string {
	return `{{define "content"}}
<div class="row">
            <div class="fixed-bottom" style="z-index: 1; left: auto; right: 30px; bottom: 20px;">
                <a href="#" data-toggle="modal" data-target="#modal-streams-add">
                    <span class="fa-stack fa-2x">
                      <i class="fas fa-circle fa-stack-2x" style="color:DodgerBlue"></i>
                      <i class="fas fa-plus fa-stack-1x fa-inverse "></i>
                    </span>
                </a>
            </div>
           
            <div class="col">
                <div id="toolbar-streams">
                    <button type="button" class="btn btn-primary btn-test">TEST</button>
                </div>
                <table  id="table-streams"
                        data-toggle="table"
                        data-toolbar="#toolbar-streams"
                        data-search="true"
                        data-url="/streams"
                        data-pagination="true"
                        data-side-pagination="client"
                        data-show-refresh="true"
						data-show-columns="true"
                        data-sort-name="uri"
                        data-sort-order="asc">
                    <thead>
                    <tr>
                        <th data-formatter="streamsControlFormatter" data-events="streamsActiveEvents">Control</th>
                        <th data-field="id">Stream-ID</th>
                        <th data-field="name">Name</th>
                        <th data-field="status">Status</th>
                        <th data-field="uri" data-visible="false">URI</th>
                        <th data-field="enabled">Auto Start</th>
                        <th data-field="urlHash">urlHash</th>
                        <th data-field="recording" data-formatter="streamsRecordingFormatter">Recording</th>
                        <th data-field="created" data-formatter="streamsCreatedFormatter" data-visible="false">Created</th>
                        <th data-field="updated" data-formatter="streamsUpdatedFormatter">Updated</th>
                    </tr>
                    </thead>
                </table>
            </div>
        </div>

        <!-- Modal -->
        <form id="form-streams-add">
            <div class="modal fade modal-form" id="modal-streams-add" tabindex="-1" role="dialog" aria-hidden="true">
                <div class="modal-dialog" role="document">
                    <div class="modal-content">
                        <div class="modal-header">
                            <h5 class="modal-title">Modal title</h5>
                            <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                                <span aria-hidden="true">&times;</span>
                            </button>
                        </div>
                        <div class="modal-body">
                            <div class="form-group">
                                <label class="form-label">Name</label>
                                <input type="text" name="name" class="form-control" value="New Stream" />
                            </div>
                            <div class="form-group">
                                <label class="form-label">URI</label>
                                <input type="text" name="uri" class="form-control" value="rtsp://10.0.75.1:8801/MainCam" />
                            </div>

                            <div class="row">
                                <div class="col">
                                    <div class="form-group">
                                        <label class="form-label">Username</label>
                                        <input type="text" name="username" class="form-control" value="admin" />
                                    </div>

                                    <div class="form-group">
                                        <label class="form-label">Password</label>
                                        <input type="password" name="password" class="form-control" value="1234"/>
                                    </div>
                                </div>
                                <div class="col">
                                    <div class="custom-control custom-switch">
                                        <input type="checkbox" name="recording" class="custom-control-input" id="customSwitchAddRecording">
                                        <label class="custom-control-label" for="customSwitchAddRecording">Recording</label>
                                    </div>

                                    <div class="custom-control custom-switch">
                                        <input type="checkbox" name="enabled" class="custom-control-input" id="customSwitchAddEnabled" checked>
                                        <label class="custom-control-label" for="customSwitchAddEnabled">Auto Start</label>
                                    </div>
                                </div>
                            </div>

                            <div class="alert alert-danger d-none" role="alert">
                                <strong>Error!</strong> <span class="msg"></span>
                            </div>

                        </div>
                        <div class="modal-footer">
                            <button type="button" class="btn btn-primary btn-streams-add">Add</button>
                            <button type="button" class="btn btn-secondary" data-dismiss="modal">Cancel</button>
                        </div>
                    </div>
                </div>
            </div>
        </form>

        <!-- Modal -->
        <form id="form-streams-edit">
            <div class="modal fade modal-form" id="modal-streams-edit" tabindex="-1" role="dialog" aria-hidden="true">
                <div class="modal-dialog" role="document">
                    <div class="modal-content">
                        <div class="modal-header">
                            <h5 class="modal-title">Edit</h5>
                            <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                                <span aria-hidden="true">&times;</span>
                            </button>
                        </div>
                        <div class="modal-body">
                            <div class="form-group">
                                <label class="form-label">Name</label>
                                <input type="text" name="name" class="form-control"/>
                            </div>
                            <div class="form-group">
                                <label class="form-label">URI</label>
                                <input type="text" name="uri" class="form-control"/>
                            </div>

                            <div class="row">
                                <div class="col">
                                    <div class="form-group">
                                        <label class="form-label">Username</label>
                                        <input type="text" name="username" class="form-control"/>
                                    </div>

                                    <div class="form-group">
                                        <label class="form-label">Password</label>
                                        <input type="password" name="password" class="form-control"/>
                                    </div>
                                </div>
                                <div class="col">
                                    <div class="custom-control custom-switch">
                                        <input type="checkbox" name="recording" class="custom-control-input" id="customSwitchEditRecording">
                                        <label class="custom-control-label" for="customSwitchEditRecording">Recording</label>
                                    </div>

                                    <div class="custom-control custom-switch">
                                        <input type="checkbox" name="enabled" class="custom-control-input" id="customSwitchEditEnabled">
                                        <label class="custom-control-label" for="customSwitchEditEnabled">Auto Start</label>
                                    </div>
                                </div>
                            </div>

                            <div class="alert alert-danger d-none" role="alert">
                                <strong>Error!</strong> <span class="msg"></span>
                            </div>

                        </div>
                        <div class="modal-footer">
                            <button type="button" class="btn btn-primary btn-streams-update">Update</button>
                            <button type="button" class="btn btn-secondary" data-dismiss="modal">Cancel</button>
                        </div>
                    </div>
                </div>
            </div>
        </form>
{{end}}

{{define "script"}}
	<script src="/static/assets/modules/stream/formatter.js"></script>
	<script src="/static/assets/modules/stream/streams.js"></script>
{{end}}
`
}

//
//
//func DisplayStream(w http.ResponseWriter, r *http.Request) {
//	tmpl := template.New("streams")
//
//	tmpl, err := tmpl.Parse(Base(Fluid))
//	if err != nil {
//		//server.Response(w, r, err, http.StatusOK)
//	}
//	if tmpl, err = tmpl.Parse(StreamPage()); err != nil {
//		//server.Response(w, r, err, http.StatusOK)
//	}
//	tmpl.Execute(w, nil)
//
//	//var err error
//	//if tmpl, err = tmpl.Parse(page); err != nil {
//	//	fmt.Println(err)
//	//}
//	//if tmpl, err = tmpl.Parse(tags); err != nil {
//	//	fmt.Println(err)
//	//}
//	//if tmpl, err = tmpl.Parse(comment); err != nil {
//	//	fmt.Println(err)
//	//}
//
//	//tmpl.Execute(os.Stdout, pagedata)
//}
