package streaming

import (
	"github.com/ghodss/yaml"
	"io/ioutil"
)

type Config struct {
	Storage struct {
		Address   string
		AccessKey string
		SecretKey string
		Bucket    string
		UseSSL    bool
		Live      string
	}
	BindAddress string `json:"bind-address"`
	Timezone    string
	Static      struct {
		Dir string
	}
}

func ReadConfig(path string) (*Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal(b, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
