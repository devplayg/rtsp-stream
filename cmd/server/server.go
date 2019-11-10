package main

import (
	"github.com/devplayg/hippo"
	"github.com/devplayg/rtsp-stream/server"
	"github.com/sirupsen/logrus"
)

const (
	appName        = "rtsp-server"
	appDisplayName = "RTSP Stream Server"
	appDescription = "RTSP Stream Server"
	appVersion     = "1.0.0"
)

func main() {
	config := &hippo.Config{
		Name:        appName,
		DisplayName: appDisplayName,
		Description: appDescription,
		Version:     appVersion,
		Debug:       true,
	}

	server := &server.Server{}
	engine := hippo.NewEngine(server, config)
	err := engine.Start()
	if err != nil {
		logrus.Fatal(err)
	}
}
