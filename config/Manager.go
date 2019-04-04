package config

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"github.com/opwire/opwire-agent/invokers"
	"github.com/opwire/opwire-agent/storages"
)

type Configuration struct {
	Version string `json:"version"`
	Agent *ConfigAgent `json:"agent"`
	Main *invokers.CommandEntrypoint `json:"main-resource"`
	Resources map[string]invokers.CommandEntrypoint `json:"resources"`
	Settings map[string]interface{} `json:"settings"`
	SettingsFormat *string `json:"settings-format"`
	HttpServer *ConfigHttpServer `json:"http-server"`
}

type ConfigAgent struct {
	ExplanationEnabled *bool `json:"explanation-enabled"`
}

type ConfigHttpServer struct {
	Host *string `json:"host"`
	Port *uint `json:"port"`
	BaseUrl *string `json:"baseurl"`
	concurrentLimitEnabled *bool `json:"concurrent-limit-enabled"`
	concurrentLimitTotal *int `json:"concurrent-limit-total"`
	singleFlightEnabled *bool `json:"single-flight-enabled"`
}

type Manager struct {
	currentVersion string
	defaultCfgFile string
	locator *Locator
	validator *Validator
}

func NewManager(currentVersion string, defaultCfgFile string) (*Manager) {
	m := &Manager{}
	m.currentVersion = currentVersion
	m.defaultCfgFile = defaultCfgFile
	m.locator = NewLocator()
	m.validator = NewValidator()
	return m
}

func (m *Manager) Load() (cfg *Configuration, result ValidationResult, err error) {
	cfg, err = m.loadJson()
	if cfg == nil || err != nil {
		return nil, nil, err
	}
	result, err = m.validator.Validate(cfg)
	return cfg, result, err
}

func (m *Manager) loadJson() (*Configuration, error) {
	fs := storages.GetFs()
	cfgpath, from := m.locator.GetConfigPath(m.defaultCfgFile)
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
	err = parser.Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (m *Manager) Init(cfg *Configuration, result ValidationResult, err error) (*Configuration, error) {
	if err != nil {
		return cfg, err
	}
	if result != nil && !result.Valid() {
		errstrs := []string {"The configuration is invalid. Errors:"}
		for _, desc := range result.Errors() {
			errstrs = append(errstrs, fmt.Sprintf("%s", desc))
		}
		return cfg, fmt.Errorf(strings.Join(errstrs, "\n - "))
	}
	if cfg != nil {
		
	}
	return cfg, nil
}

func (c *ConfigHttpServer) ConcurrentLimitEnabled() bool {
	if c.concurrentLimitEnabled == nil {
		return false
	}
	return *c.concurrentLimitEnabled
}

func (c *ConfigHttpServer) ConcurrentLimitTotal() int {
	if c.concurrentLimitTotal == nil {
		return 0
	}
	return *c.concurrentLimitTotal
}

func (c *ConfigHttpServer) SingleFlightEnabled() bool {
	if c.singleFlightEnabled == nil {
		return false
	}
	return *c.singleFlightEnabled
}
