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
	defaultFile string
	fs afero.Fs
}

func NewLoader(defaultCfgFile string) (*Loader) {
	l := &Loader{}
	l.defaultFile = defaultCfgFile
	l.fs = afero.NewOsFs()
	return l
}

func (l *Loader) Load() (*Configuration, error) {
	config := &Configuration{}
	configFile, err := l.fs.Open(l.defaultFile)
	defer configFile.Close()
	if err != nil {
		return nil, err
	}
	parser := json.NewDecoder(configFile)
	parser.Decode(config)
	return config, nil
}
