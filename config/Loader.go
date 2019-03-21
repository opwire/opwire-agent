package config

import (
	"encoding/json"
	"github.com/opwire/opwire-agent/invokers"
	"github.com/spf13/afero"
)

type Configuration struct {
	Version string `json:"version"`
	Main invokers.CommandEntrypoint `json:"main-resource"`
	Resources map[string]invokers.CommandEntrypoint `json:"resources"`
	Unformed map[string]interface{} `json:"unformed"`
}

type Loader struct {
	fs afero.Fs
}

func NewLoader() (*Loader) {
	l := &Loader{}
	l.fs = afero.NewOsFs()
	return l
}

func (l *Loader) Load(file string) (*Configuration, error) {
	config := &Configuration{}
	configFile, err := l.fs.Open(file)
	defer configFile.Close()
	if err != nil {
		return nil, err
	}
	parser := json.NewDecoder(configFile)
	parser.Decode(config)
	return config, nil
}
