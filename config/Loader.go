package config

import (
	"encoding/json"
	"log"
	"github.com/opwire/opwire-agent/invokers"
	"github.com/opwire/opwire-agent/storages"
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
}

func NewLoader(defaultCfgFile string) (*Loader) {
	l := &Loader{}
	l.defaultFile = defaultCfgFile
	l.locator = NewLocator()
	return l
}

func (l *Loader) Load() (*Configuration, error) {
	fs := storages.GetFs()
	cfgpath, from := l.locator.GetConfigPath(l.defaultFile)
	if len(from) == 0 {
		log.Printf("Configuration file not found")
		return nil, nil
	} else {
		log.Printf("Configuration path [%s] from [%s]", cfgpath, from)
	}

	config := &Configuration{}
	configFile, err := fs.Open(cfgpath)
	defer configFile.Close()
	if err != nil {
		return nil, err
	}
	parser := json.NewDecoder(configFile)
	parser.Decode(config)
	return config, nil
}
