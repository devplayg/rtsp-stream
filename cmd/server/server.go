package main

import (
	"github.com/devplayg/hippo"
	"github.com/devplayg/rtsp-stream/streaming"
	log "github.com/sirupsen/logrus"
)

const (
	appName        = "rtsp-server"
	appDisplayName = "RTSP Stream"
	appDescription = "RTSP Stream"
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

	configPath := "config.yaml"
	appConfig, err := streaming.ReadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	hippo.InitLogger("", appName, config.Debug, config.Verbose)

	server := streaming.NewServer(appConfig.BindAddress, appConfig.Storage.Live, appConfig.Storage.Recording)
	engine := hippo.NewEngine(server, config)
	if err := engine.Start(); err != nil {
		log.Fatal(err)
	}
}
