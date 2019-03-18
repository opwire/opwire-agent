package config

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	Version string `json:"version"`
	Unformed map[string]interface{} `json:"unformed"`
	Mappings map[string]CommandDefinition `json:"mappings"`
}

type CommandDefinition struct {
	GlobalCommand string `json:"command"`
	MethodCommand map[string]string `json:"methods"`
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
