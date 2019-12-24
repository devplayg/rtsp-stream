package server

import (
	"github.com/devplayg/eggcrate"
)

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
	/assets/plugins/moment/moment-timezone-with-data.min.js
	/assets/plugins/moment/moment-timezone.min.js
	/assets/plugins/moment/moment.min.js
	/assets/plugins/videojs/video-js.min.css
	/assets/plugins/videojs/video.min.js
	/assets/plugins/videojs/videojs-http-streaming.min.js
*/

var (
	uiAssetMap map[string][]byte
)

func init() {
	var err error
	uiAssetMap, err = eggcrate.Decode(assetData)
	if err != nil {
		panic(err)
	}
}
