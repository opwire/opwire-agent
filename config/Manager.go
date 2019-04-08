package config

import (
	"encoding/json"
	"log"
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
	ConcurrentLimit *SectionConcurrentLimit `json:"concurrent-limit"`
	SingleFlight *SectionSingleFlight `json:"single-flight"`
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
	return c.GetConcurrentLimit().GetEnabled()
}

func (c *ConfigHttpServer) ConcurrentLimitTotal() int {
	return c.GetConcurrentLimit().GetTotal()
}

func (c *ConfigHttpServer) SingleFlightEnabled() bool {
	return c.GetSingleFlight().GetEnabled()
}

func (c *ConfigHttpServer) SingleFlightReqIdName() string {
	return c.GetSingleFlight().GetReqIdName()
}

func (c *ConfigHttpServer) SingleFlightByMethod() bool {
	return c.GetSingleFlight().GetByMethod()
}

func (c *ConfigHttpServer) SingleFlightByPath() bool {
	return c.GetSingleFlight().GetByPath()
}

func (c *ConfigHttpServer) SingleFlightByHeaders() []string {
	return c.GetSingleFlight().GetByHeaders()
}

func (c *ConfigHttpServer) SingleFlightByQueries() []string {
	return c.GetSingleFlight().GetByQueries()
}

func (c *ConfigHttpServer) SingleFlightByBody() bool {
	return c.GetSingleFlight().GetByBody()
}

func (c *ConfigHttpServer) SingleFlightByUserIP() bool {
	return c.GetSingleFlight().GetByUserIP()
}

func (c *ConfigHttpServer) GetConcurrentLimit() *SectionConcurrentLimit {
	if c.ConcurrentLimit == nil {
		return &SectionConcurrentLimit{}
	}
	return c.ConcurrentLimit
}

type SectionConcurrentLimit struct {
	Enabled *bool `json:"enabled"`
	Total *int `json:"total"`
}

func (c *SectionConcurrentLimit) GetEnabled() bool {
	if c.Enabled == nil {
		return false
	}
	return *c.Enabled
}

func (c *SectionConcurrentLimit) GetTotal() int {
	if c.Total == nil {
		return 0
	}
	return *c.Total
}

func (c *ConfigHttpServer) GetSingleFlight() *SectionSingleFlight {
	if c.SingleFlight == nil {
		return &SectionSingleFlight{}
	}
	return c.SingleFlight
}

type SectionSingleFlight struct {
	Enabled *bool `json:"enabled"`
	ReqIdName *string `json:"req-id"`
	ByMethod *bool `json:"by-method"`
	ByPath *bool `json:"by-path"`
	ByHeaders *string `json:"by-headers"`
	ByQueries *string `json:"by-queries"`
	ByBody *bool `json:"by-body"`
	ByUserIP *bool `json:"by-userip"`
}

func (c *SectionSingleFlight) GetEnabled() bool {
	if c.Enabled == nil {
		return false
	}
	return *c.Enabled
}

func (c *SectionSingleFlight) GetReqIdName() string {
	if c.ReqIdName == nil {
		return ""
	}
	return *c.ReqIdName
}

func (c *SectionSingleFlight) GetByMethod() bool {
	if c.ByMethod == nil {
		if c.ReqIdName != nil && len(*c.ReqIdName) > 0 {
			return false
		}
		if c.ByPath != nil && *c.ByPath {
			return true
		}
		if c.ByUserIP != nil && *c.ByUserIP {
			return true
		}
		return false
	}
	return *c.ByMethod
}

func (c *SectionSingleFlight) GetByPath() bool {
	if c.ByPath == nil {
		if c.ReqIdName != nil && len(*c.ReqIdName) > 0 {
			return false
		}
		if c.ByMethod != nil && *c.ByMethod {
			return true
		}
		if c.ByUserIP != nil && *c.ByUserIP {
			return true
		}
		return false
	}
	return *c.ByPath
}

func (c *SectionSingleFlight) GetByHeaders() []string {
	if c.ByHeaders == nil {
		return []string{}
	}
	return utils.Split(*c.ByHeaders, ",")
}

func (c *SectionSingleFlight) GetByQueries() []string {
	if c.ByQueries == nil {
		return []string{}
	}
	return utils.Split(*c.ByQueries, ",")
}

func (c *SectionSingleFlight) GetByBody() bool {
	if c.ByBody == nil {
		return false
	}
	return *c.ByBody
}

func (c *SectionSingleFlight) GetByUserIP() bool {
	if c.ByUserIP == nil {
		if c.ReqIdName != nil && len(*c.ReqIdName) > 0 {
			return false
		}
		if c.ByMethod != nil && *c.ByMethod {
			return true
		}
		if c.ByPath != nil && *c.ByPath {
			return true
		}
		return false
	}
	return *c.ByUserIP
}
