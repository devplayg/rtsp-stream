package main

import (
	"github.com/devplayg/hippo"
	"github.com/devplayg/rtsp-stream/common"
	"github.com/devplayg/rtsp-stream/store"
	log "github.com/sirupsen/logrus"
)

const (
	appName        = "rtsp-alba"
	appDisplayName = "RTSP Alba"
	appDescription = "RTSP Alba"
	appVersion     = "1.0.0"
)

func main() {
	config := &hippo.Config{
		Name:        appName,
		DisplayName: appDisplayName,
		Description: appDescription,
		Version:     appVersion,
		IsService:   false,
		Debug:       true,
		Verbose:     true,
	}

	hippo.InitLogger("", appName, config.Debug, config.Verbose)
	alba := store.NewAlba(common.ReadConfig("config.yaml"))
	engine := hippo.NewEngine(alba, config)
	if err := engine.Start(); err != nil {
		log.Fatal(err)
	}
}
