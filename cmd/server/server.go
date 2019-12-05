package main

import (
	"github.com/devplayg/hippo"
	"github.com/devplayg/rtsp-stream/common"
	"github.com/devplayg/rtsp-stream/streaming"
	log "github.com/sirupsen/logrus"
)

const (
	appName        = "rtsp-server"
	appDisplayName = "RTSP Streamer"
	appDescription = "RTSP Streamer"
	appVersion     = "1.0.0"
)

func main() {
	config := &hippo.Config{
		Name:        appName,
		DisplayName: appDisplayName,
		Description: appDescription,
		Version:     appVersion,
		IsService:   true,
		Debug:       true,
		Verbose:     true,
	}
	hippo.InitLogger("", appName, config.Debug, config.Verbose)
	server := streaming.NewServer(common.ReadConfig("config.yaml"))
	engine := hippo.NewEngine(server, config)
	if err := engine.Start(); err != nil {
		log.Fatal(err)
	}
}
