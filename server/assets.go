package server

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
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

//func decodeAssets() (map[string][]byte, error) {
//	var m map[string][]byte
//	b, err := base64.StdEncoding.DecodeString(encodedUiAssets)
//	if err != nil {
//		return nil, err
//	}
//
//	err = json.Unmarshal(b, &m)
//	return m, err
//}

func init() {

	compressed, err := decodeAssets(assetData)
	if err != nil {
		panic(err)
	}

	b, err := decompress(compressed)
	if err != nil {
		panic(err)
	}

	uiAssetMap = make(map[string][]byte)
	err = json.Unmarshal(b, &uiAssetMap)

	if err != nil {
		panic(err)
	}

	for k, _ := range uiAssetMap {
		fmt.Printf("%s is loaded. len=%d\n", k, len(uiAssetMap[k]))
	}
}

func decompress(b []byte) ([]byte, error) {
	buf := bytes.NewBuffer(b)

	var r io.Reader
	var err error
	r, err = gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}

	var resB bytes.Buffer
	_, err = resB.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	return resB.Bytes(), nil

}

func decodeAssets(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}
