package common

import (
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

type Config struct {
	Storage struct {
		Remote    bool
		Address   string
		AccessKey string
		SecretKey string
		Bucket    string
		UseSSL    bool
		LiveDir   string
		RecordDir string
	} `json:"storage"`
	BindAddress string `json:"bind-address"`
	Timezone    string
	Static      struct {
		Dir string
	}
	HlsOptions struct {
		SegmentTime int
	}
}

func ReadConfig(path string) *Config {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Warn(err)
		return &DefaultConfig
	}

	config := &DefaultConfig
	err = yaml.Unmarshal(b, config)
	if err != nil {
		log.Warn(err)
		return &DefaultConfig
	}

	return config
}

var DefaultConfig = Config{
	Storage: struct {
		Remote    bool
		Address   string
		AccessKey string
		SecretKey string
		Bucket    string
		UseSSL    bool
		LiveDir   string
		RecordDir string
	}{
		Remote:    false,
		Address:   "127.0.0.1:9000",
		AccessKey: "admin",
		SecretKey: "adminpw",
		Bucket:    "record",
		UseSSL:    false,
		LiveDir:   "live",
		RecordDir: "storage",
	},
	BindAddress: "0.0.0.0:8000",
	Static: struct {
		Dir string
	}{Dir: "static"},
	HlsOptions: HlsOption{SegmentTime: 30},
}

type HlsOption struct {
	SegmentTime int
}

// wondory
//func GetDefaultConfig() *Config {
//    return &Config{
//        Storage: struct {
//            Address   string
//            AccessKey string
//            SecretKey string
//            Bucket    string
//            UseSSL    bool
//            LiveDir      string
//        }{
//            Address:   "127.0.0.1:9000",
//            AccessKey: "admin",
//            SecretKey: "admin",
//            Bucket:    "record",
//            UseSSL:    false,
//			LiveDir:      "./live",
//        },
//        BindAddress: "127.0.0.1:8000",
//        Timezone:    "",
//        Static: struct {
//            Dir string
//        }{
//            Dir: "./static",
//        },
//    }
//}
