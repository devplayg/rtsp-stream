package main

import (
	"fmt"
	"github.com/devplayg/hippo"
	"github.com/devplayg/rtsp-stream/common"
	"github.com/devplayg/rtsp-stream/streaming"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"os"
)

const (
	appName        = "rtsp-server"
	appDisplayName = "RTSP Streamer"
	appDescription = "RTSP Streamer"
	appVersion     = "1.0.0"
)

var (
	fs         = pflag.NewFlagSet(appName, pflag.ContinueOnError)
	debug      = fs.Bool("debug", true, "Debug")
	verbose    = fs.BoolP("verbose", "v", false, "Verbose")
	version    = fs.Bool("version", false, "Version")
	configPath = fs.StringP("config", "c", "config.yaml", "Configuration file")
)

func main() {
	config := &hippo.Config{
		Name:        appName,
		DisplayName: appDisplayName,
		Description: appDescription,
		Version:     appVersion,
		Debug:       *debug,
		Verbose:     *verbose,
		IsService:   true,
	}

	server := streaming.NewServer(common.ReadConfig(*configPath))
	engine := hippo.NewEngine(server, config)
	if err := engine.Start(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	fs.Usage = hippo.Usage(fs, appDisplayName, appVersion)
	_ = fs.Parse(os.Args[1:])

	if *version {
		fmt.Printf("%s %s\n", appDisplayName, appVersion)
		os.Exit(1)
	}
	hippo.InitLogger("", appName, *debug, *verbose)
}
