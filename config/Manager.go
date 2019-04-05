package config

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"github.com/opwire/opwire-agent/invokers"
	"github.com/opwire/opwire-agent/storages"
	"github.com/opwire/opwire-agent/utils"
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
	MyConcurrentLimitEnabled *bool `json:"concurrent-limit-enabled"`
	MyConcurrentLimitTotal *int `json:"concurrent-limit-total"`
	MySingleFlightEnabled *bool `json:"single-flight-enabled"`
	MySingleFlightReqIdName *string `json:"single-flight-req-id"`
	MySingleFlightByMethod *bool `json:"single-flight-by-method"`
	MySingleFlightByPath *bool `json:"single-flight-by-path"`
	MySingleFlightByHeaders *string `json:"single-flight-by-headers"`
	MySingleFlightByQueries *string `json:"single-flight-by-queries"`
	MySingleFlightByBody *bool `json:"single-flight-by-body"`
	MySingleFlightByUserIP *bool `json:"single-flight-by-userip"`
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
	if c.MyConcurrentLimitEnabled == nil {
		return false
	}
	return *c.MyConcurrentLimitEnabled
}

func (c *ConfigHttpServer) ConcurrentLimitTotal() int {
	if c.MyConcurrentLimitTotal == nil {
		return 0
	}
	return *c.MyConcurrentLimitTotal
}

func (c *ConfigHttpServer) SingleFlightEnabled() bool {
	if c.MySingleFlightEnabled == nil {
		return true
	}
	return *c.MySingleFlightEnabled
}

func (c *ConfigHttpServer) SingleFlightReqIdName() string {
	if c.MySingleFlightReqIdName == nil {
		return ""
	}
	return *c.MySingleFlightReqIdName
}

func (c *ConfigHttpServer) SingleFlightByMethod() bool {
	if c.MySingleFlightByMethod == nil {
		return true
	}
	return *c.MySingleFlightByMethod
}

func (c *ConfigHttpServer) SingleFlightByPath() bool {
	if c.MySingleFlightByPath == nil {
		return true
	}
	return *c.MySingleFlightByPath
}

func (c *ConfigHttpServer) SingleFlightByHeaders() []string {
	if c.MySingleFlightByHeaders == nil {
		return []string{}
	}
	return utils.Split(*c.MySingleFlightByHeaders, ",")
}

func (c *ConfigHttpServer) SingleFlightByQueries() []string {
	if c.MySingleFlightByQueries == nil {
		return []string{}
	}
	return utils.Split(*c.MySingleFlightByQueries, ",")
}

func (c *ConfigHttpServer) SingleFlightByBody() bool {
	if c.MySingleFlightByBody == nil {
		return false
	}
	return *c.MySingleFlightByBody
}

func (c *ConfigHttpServer) SingleFlightByUserIP() bool {
	if c.MySingleFlightByUserIP == nil {
		return false
	}
	return *c.MySingleFlightByUserIP
}
