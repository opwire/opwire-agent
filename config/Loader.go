package config

import (
	"encoding/json"
	"log"
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
	locator *Locator
	fs afero.Fs
}

func NewLoader(defaultCfgFile string) (*Loader) {
	l := &Loader{}
	l.defaultFile = defaultCfgFile
	l.locator = &Locator{}
	l.assignFs(afero.NewOsFs())
	return l
}

func (l *Loader) assignFs(newFs afero.Fs) {
	l.locator.fs = newFs
	l.fs = newFs
}

func (l *Loader) Load() (*Configuration, error) {
	cfgpath, from := l.locator.GetConfigPath(l.defaultFile)
	if len(from) == 0 {
		log.Printf("Configuration file not found")
		return nil, nil
	} else {
		log.Printf("Configuration path [%s] from [%s]", cfgpath, from)
	}

	config := &Configuration{}
	configFile, err := l.fs.Open(cfgpath)
	defer configFile.Close()
	if err != nil {
		return nil, err
	}
	parser := json.NewDecoder(configFile)
	parser.Decode(config)
	return config, nil
}
