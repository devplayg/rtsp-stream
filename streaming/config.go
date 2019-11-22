package streaming

import (
	"github.com/ghodss/yaml"
	"io/ioutil"
)

type Config struct {
	Storage struct {
		Live      string
		Recording string
	}
	BindAddress string `json:"bind-address"`
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
