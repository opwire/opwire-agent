package config

import (
	"encoding/json"
	"os"
	"github.com/opwire/opwire-agent/invokers"
)

type Configuration struct {
	Version string `json:"version"`
	Resources map[string]invokers.CommandEntrypoint `json:"resources"`
	Unformed map[string]interface{} `json:"unformed"`
}

type Loader struct {}

func NewLoader() (*Loader) {
	l := &Loader{}
	return l
}

func (m *Loader) Load(file string) (*Configuration, error) {
	config := &Configuration{}
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		return nil, err
	}
	parser := json.NewDecoder(configFile)
	parser.Decode(config)
	return config, nil
}
