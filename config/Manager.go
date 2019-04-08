package config

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
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
	OutputCombined *bool `json:"combine-stderr-stdout"` // 2>&1
}

type ConfigHttpServer struct {
	managerOptions ManagerOptions
	Host *string `json:"host"`
	Port *uint `json:"port"`
	MaxHeaderBytes *int `json:"max-header-bytes"`
	ReadTimeout *string `json:"read-timeout"`
	WriteTimeout *string `json:"write-timeout"`
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
	locator *Locator
	validator *Validator
	options ManagerOptions
}

type ManagerOptions interface {
	GetConfigPath() string
	GetHost() string
	GetPort() uint
	GetVersion() string
}

func NewManager(options ManagerOptions) (*Manager) {
	m := &Manager{}
	m.locator = NewLocator()
	m.validator = NewValidator()
	m.options = options
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
	cfgpath, from := m.locator.GetConfigPath(m.options.GetConfigPath())
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

	if config.HttpServer != nil {
		config.HttpServer.managerOptions = m.options
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

func (c *ConfigHttpServer) GetHost() string {
	o := c.managerOptions
	if o != nil && o.GetHost() != "" {
		return o.GetHost()
	}
	if c.Host != nil {
		return *c.Host
	}
	return ""
}

func (c *ConfigHttpServer) GetPort() uint {
	o := c.managerOptions
	if o != nil && o.GetPort() != 0 {
		return o.GetPort()
	}
	if c.Port != nil {
		return *c.Port
	}
	return 0
}

func (c *ConfigHttpServer) GetBaseUrl() string {
	if c.BaseUrl != nil {
		return *c.BaseUrl
	}
	return ""
}

func (c *ConfigHttpServer) GetMaxHeaderBytes() int {
	if c.MaxHeaderBytes != nil {
		return *c.MaxHeaderBytes
	}
	return 0
}

func (c *ConfigHttpServer) GetReadTimeout() (time.Duration, error) {
	if c.ReadTimeout != nil {
		return time.ParseDuration(*c.ReadTimeout)
	}
	return 0, nil
}

func (c *ConfigHttpServer) GetWriteTimeout() (time.Duration, error) {
	if c.WriteTimeout != nil {
		return time.ParseDuration(*c.WriteTimeout)
	}
	return 0, nil
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
		if c.MySingleFlightReqIdName != nil && len(*c.MySingleFlightReqIdName) > 0 {
			return false
		}
		if c.MySingleFlightByPath != nil && *c.MySingleFlightByPath {
			return true
		}
		if c.MySingleFlightByUserIP != nil && *c.MySingleFlightByUserIP {
			return true
		}
		return false
	}
	return *c.MySingleFlightByMethod
}

func (c *ConfigHttpServer) SingleFlightByPath() bool {
	if c.MySingleFlightByPath == nil {
		if c.MySingleFlightReqIdName != nil && len(*c.MySingleFlightReqIdName) > 0 {
			return false
		}
		if c.MySingleFlightByMethod != nil && *c.MySingleFlightByMethod {
			return true
		}
		if c.MySingleFlightByUserIP != nil && *c.MySingleFlightByUserIP {
			return true
		}
		return false
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
		if c.MySingleFlightReqIdName != nil && len(*c.MySingleFlightReqIdName) > 0 {
			return false
		}
		if c.MySingleFlightByMethod != nil && *c.MySingleFlightByMethod {
			return true
		}
		if c.MySingleFlightByPath != nil && *c.MySingleFlightByPath {
			return true
		}
		return false
	}
	return *c.MySingleFlightByUserIP
}
