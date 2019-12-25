package server

import (
	"net/http"
)

func (c *Controller) initRouter() {
	c.setApiRoutes()
	c.setAssetRoutes()
	c.setUiRoutes()
}

func (c *Controller) setUiRoutes() {
	c.router.HandleFunc("/streams/", c.DisplayStreams).Methods("GET")
	c.router.HandleFunc("/videos/", c.DisplayVideos).Methods("GET")
	c.router.HandleFunc("/live/", c.DisplayLive).Methods("GET")
	c.router.HandleFunc("/tpl", serveTemplate2)
}

func (c *Controller) setApiRoutes() {
	//r.HandleFunc("/test", c.Test).Methods("GET")
	c.router.HandleFunc("/streams", c.GetStreams).Methods("GET")
	c.router.HandleFunc("/streams", c.AddStream).Methods("POST")
	c.router.HandleFunc("/streams/debug", c.DebugStream).Methods("GET")

	c.router.HandleFunc("/streams/{id:[0-9]+}", c.GetStreamById).Methods("GET")
	c.router.HandleFunc("/streams/{id:[0-9]+}", c.UpdateStream).Methods("PATCH")
	c.router.HandleFunc("/streams/{id:[0-9]+}", c.DeleteStream).Methods("DELETE")

	c.router.HandleFunc("/streams/{id:[0-9]+}/start", c.StartStream).Methods("GET")
	c.router.HandleFunc("/streams/{id:[0-9]+}/stop", c.StopStream).Methods("GET")

	// Video records
	c.router.HandleFunc("/videos", c.GetVideoRecords).Methods("GET")

	// Today M3u8: http://127.0.0.1:8000/videos/1/today/m3u8
	c.router.HandleFunc("/videos/{id:[0-9]+}/today/m3u8", c.GetTodayM3u8).Methods("GET")
	// Today videos: http://127.0.0.1:8000/videos/1/today/media0.ts
	c.router.HandleFunc("/videos/{id:[0-9]+}/today/{media}.ts", c.GetTodayVideo).Methods("GET")

	// (O) Live M3u8: http://127.0.0.1:8000/videos/1/live/m3u8
	c.router.HandleFunc("/videos/{id:[0-9]+}/live/m3u8", c.GetLiveM3u8).Methods("GET")
	// (O) Live videos: http://127.0.0.1:8000/videos/1/live/media0.ts
	c.router.HandleFunc("/videos/{id:[0-9]+}/live/{media}.ts", c.GetLiveVideo).Methods("GET")

	// Old M3u8: http://127.0.0.1:8000/videos/1/date/20191211/m3u8
	c.router.HandleFunc("/videos/{id:[0-9]+}/date/{date:[0-9]+}/m3u8", c.GetDailyM3u8).Methods("GET")
	// Old videos: http://127.0.0.1:8000/videos/1/date/20191211/media0.ts
	c.router.HandleFunc("/videos/{id:[0-9]+}/date/{date:[0-9]+}/{media}.ts", c.GetDailyVideo).Methods("GET")

	c.router.
		PathPrefix("/static").
		Handler(http.StripPrefix("/static", http.FileServer(http.Dir(c.staticDir))))
}

func (c *Controller) setAssetRoutes() {
	/*
		/assets/css/custom.js
		/assets/img/logo.png
		/assets/js/custom.js
		/assets/js/jquery-3.4.1.min.js
		/assets/js/jquery.mask.min.js
		/assets/js/js.cookie-2.2.1.min.js
		/assets/js/popper.min.js
		/assets/plugins/bootstrap-table/bootstrap-table.min.css
		-
		/assets/plugins/bootstrap/bootstrap.min.css
		/assets/plugins/bootstrap/bootstrap.min.js
		/assets/plugins/moment/moment-timezone-with-data.min.js
		/assets/plugins/moment/moment-timezone.min.js
		/assets/plugins/moment/moment.min.js
	*/

	c.router.HandleFunc("/assets/{assetType}/{name}", func(w http.ResponseWriter, r *http.Request) {
		GetAsset(w, r)
	})

	c.router.HandleFunc("/assets/plugins/{pluginName}/{name}", func(w http.ResponseWriter, r *http.Request) {
		GetAsset(w, r)
	})
	c.router.HandleFunc("/assets/plugins/{pluginName}/{kind}/{name}", func(w http.ResponseWriter, r *http.Request) {
		GetAsset(w, r)
	})

	c.router.HandleFunc("/assets/modules/{moduleName}/{name}", func(w http.ResponseWriter, r *http.Request) {
		GetAsset(w, r)
	})
}
