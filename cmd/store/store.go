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
		IsService:   true,
		Debug:       true,
		Verbose:     true,
	}

	appConfig, err := common.ReadConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	hippo.InitLogger("", appName, config.Debug, config.Verbose)
	alba := store.NewAlba(appConfig)
	engine := hippo.NewEngine(alba, config)
	if err := engine.Start(); err != nil {
		log.Fatal(err)
	}
}
