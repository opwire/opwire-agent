package config

import (
	"encoding/json"
	"log"
	"github.com/opwire/opwire-agent/invokers"
	"github.com/opwire/opwire-agent/storages"
)

type Configuration struct {
	Version string `json:"version"`
	Main *invokers.CommandEntrypoint `json:"main-resource"`
	Resources map[string]invokers.CommandEntrypoint `json:"resources"`
	Unformed map[string]interface{} `json:"unformed"`
}

type Loader struct {
	currentVersion string
	defaultCfgFile string
	locator *Locator
	validator *Validator
}

func NewLoader(currentVersion string, defaultCfgFile string) (*Loader) {
	l := &Loader{}
	l.currentVersion = currentVersion
	l.defaultCfgFile = defaultCfgFile
	l.locator = NewLocator()
	l.validator = NewValidator()
	return l
}

func (l *Loader) Load() (cfg *Configuration, result ValidationResult, err error) {
	cfg, err = l.loadJson()
	if cfg == nil || err != nil {
		return nil, nil, err
	}
	result, err = l.validator.Validate(cfg)
	return cfg, result, err
}

func (l *Loader) loadJson() (*Configuration, error) {
	fs := storages.GetFs()
	cfgpath, from := l.locator.GetConfigPath(l.defaultCfgFile)
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
