package config

import (
	"encoding/json"
	"time"
	"github.com/opwire/opwire-agent/invokers"
	"github.com/opwire/opwire-agent/storages"
	"github.com/opwire/opwire-agent/utils"
)

type Configuration struct {
	Version string `json:"version"`
	Agent *configAgent `json:"agent"`
	Main *invokers.CommandEntrypoint `json:"main-resource"`
	Resources map[string]invokers.CommandEntrypoint `json:"resources"`
	Settings map[string]interface{} `json:"settings"`
	SettingsFormat *string `json:"settings-format"`
	HttpServer *configHttpServer `json:"http-server"`
	Logging *configLogging `json:"logging"`
	managerOptions ManagerOptions
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

func (m *Manager) Load() (cfg *Configuration, result *LoadResult, err error) {
	result = &LoadResult{}
	cfg, result.configPath, result.configFrom, err = m.loadJson()
	if err != nil {
		return nil, nil, err
	}
	if cfg != nil {
		result.validationResult, err = m.validator.Validate(cfg)
	} else {
		v := "0.0.0"
		if m.options != nil {
			v = m.options.GetVersion()
		}
		cfg = &Configuration{
			Version: v,
		}
	}

	cfg.managerOptions = m.options

	return cfg, result, err
}

func (m *Manager) loadJson() (*Configuration, string, string, error) {
	fs := storages.GetFs()
	cfgpath, cfgfrom := m.locator.GetConfigPath(m.options.GetConfigPath())
	if len(cfgfrom) == 0 {
		return nil, cfgpath, cfgfrom, nil
	}

	config := &Configuration{}

	configFile, err := fs.Open(cfgpath)
	if configFile != nil {
		defer configFile.Close()
	}
	if err != nil {
		return nil, cfgpath, cfgfrom, err
	}

	parser := json.NewDecoder(configFile)
	if parser != nil {
		err = parser.Decode(config)
		if err != nil {
			return nil, cfgpath, cfgfrom, err
		}
	}

	return config, cfgpath, cfgfrom, nil
}

type LoadResult struct {
	validationResult *ValidationResult
	configFrom string
	configPath string
}

func (r *LoadResult) Valid() bool {
	if r.validationResult == nil {
		return true
	}
	return r.validationResult.Valid()
}

func (r *LoadResult) Errors() []ValidationError {
	if r.validationResult == nil {
		return nil
	}
	return r.validationResult.Errors()
}

func (r *LoadResult) GetConfigPath() string {
	return r.configPath
}

func (r *LoadResult) GetConfigFrom() string {
	return r.configFrom
}

type configAgent struct {
	Explanation *sectionExplanation `json:"explanation"`
	OutputCombined *bool `json:"combine-stderr-stdout"` // 2>&1
}

func (c *Configuration) GetAgent() *configAgent {
	if c.Agent == nil {
		return &configAgent{}
	}
	return c.Agent
}

func (c *configAgent) GetExplanation() *sectionExplanation {
	if c.Explanation == nil {
		return &sectionExplanation{}
	}
	return c.Explanation
}

func (c *configAgent) GetOutputCombined() bool {
	if c.OutputCombined == nil {
		return false
	}
	return *c.OutputCombined
}

type sectionExplanation struct {
	Enabled *bool `json:"enabled"`
	Format *string `json:"format"`
}

func (c *sectionExplanation) GetEnabled() bool {
	if c.Enabled == nil {
		return false
	}
	return *c.Enabled
}

func (c *sectionExplanation) GetFormat() string {
	if c.Format == nil {
		return ""
	}
	return *c.Format
}

type configHttpServer struct {
	managerOptions ManagerOptions
	Host *string `json:"host"`
	Port *uint `json:"port"`
	MaxHeaderBytes *int `json:"max-header-bytes"`
	ReadTimeout *string `json:"read-timeout"`
	WriteTimeout *string `json:"write-timeout"`
	BaseUrl *string `json:"baseurl"`
	ConcurrentLimit *sectionConcurrentLimit `json:"concurrent-limit"`
	SingleFlight *sectionSingleFlight `json:"single-flight"`
}

func (c *Configuration) GetHttpServer() *configHttpServer {
	httpServer := c.HttpServer
	if httpServer == nil {
		httpServer = &configHttpServer{}
	}
	if httpServer.managerOptions == nil {
		httpServer.managerOptions = c.managerOptions
	}
	return httpServer
}

func (c *configHttpServer) GetHost() string {
	o := c.managerOptions
	if o != nil && o.GetHost() != "" {
		return o.GetHost()
	}
	if c.Host != nil {
		return *c.Host
	}
	return ""
}

func (c *configHttpServer) GetPort() uint {
	o := c.managerOptions
	if o != nil && o.GetPort() != 0 {
		return o.GetPort()
	}
	if c.Port != nil {
		return *c.Port
	}
	return 0
}

func (c *configHttpServer) GetBaseUrl() string {
	if c.BaseUrl != nil {
		return *c.BaseUrl
	}
	return ""
}

func (c *configHttpServer) GetMaxHeaderBytes() int {
	if c.MaxHeaderBytes != nil {
		return *c.MaxHeaderBytes
	}
	return 0
}

func (c *configHttpServer) GetReadTimeout() (time.Duration, error) {
	if c.ReadTimeout != nil {
		return time.ParseDuration(*c.ReadTimeout)
	}
	return 0, nil
}

func (c *configHttpServer) GetWriteTimeout() (time.Duration, error) {
	if c.WriteTimeout != nil {
		return time.ParseDuration(*c.WriteTimeout)
	}
	return 0, nil
}

func (c *configHttpServer) ConcurrentLimitEnabled() bool {
	return c.GetConcurrentLimit().GetEnabled()
}

func (c *configHttpServer) ConcurrentLimitTotal() int {
	return c.GetConcurrentLimit().GetTotal()
}

func (c *configHttpServer) SingleFlightEnabled() bool {
	return c.GetSingleFlight().GetEnabled()
}

func (c *configHttpServer) SingleFlightReqIdName() string {
	return c.GetSingleFlight().GetReqIdName()
}

func (c *configHttpServer) SingleFlightByMethod() bool {
	return c.GetSingleFlight().GetByMethod()
}

func (c *configHttpServer) SingleFlightByPath() bool {
	return c.GetSingleFlight().GetByPath()
}

func (c *configHttpServer) SingleFlightByHeaders() []string {
	return c.GetSingleFlight().GetByHeaders()
}

func (c *configHttpServer) SingleFlightByQueries() []string {
	return c.GetSingleFlight().GetByQueries()
}

func (c *configHttpServer) SingleFlightByBody() bool {
	return c.GetSingleFlight().GetByBody()
}

func (c *configHttpServer) SingleFlightByUserIP() bool {
	return c.GetSingleFlight().GetByUserIP()
}

func (c *configHttpServer) GetConcurrentLimit() *sectionConcurrentLimit {
	if c.ConcurrentLimit == nil {
		return &sectionConcurrentLimit{}
	}
	return c.ConcurrentLimit
}

type sectionConcurrentLimit struct {
	Enabled *bool `json:"enabled"`
	Total *int `json:"total"`
}

func (c *sectionConcurrentLimit) GetEnabled() bool {
	if c.Enabled == nil {
		return false
	}
	return *c.Enabled
}

func (c *sectionConcurrentLimit) GetTotal() int {
	if c.Total == nil {
		return 0
	}
	return *c.Total
}

func (c *configHttpServer) GetSingleFlight() *sectionSingleFlight {
	if c.SingleFlight == nil {
		return &sectionSingleFlight{}
	}
	return c.SingleFlight
}

type sectionSingleFlight struct {
	Enabled *bool `json:"enabled"`
	ReqIdName *string `json:"req-id"`
	ByMethod *bool `json:"by-method"`
	ByPath *bool `json:"by-path"`
	ByHeaders *string `json:"by-headers"`
	ByQueries *string `json:"by-queries"`
	ByBody *bool `json:"by-body"`
	ByUserIP *bool `json:"by-userip"`
}

func (c *sectionSingleFlight) GetEnabled() bool {
	if c.Enabled == nil {
		return false
	}
	return *c.Enabled
}

func (c *sectionSingleFlight) GetReqIdName() string {
	if c.ReqIdName == nil {
		return ""
	}
	return *c.ReqIdName
}

func (c *sectionSingleFlight) GetByMethod() bool {
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

func (c *sectionSingleFlight) GetByPath() bool {
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

func (c *sectionSingleFlight) GetByHeaders() []string {
	if c.ByHeaders == nil {
		return []string{}
	}
	return utils.Split(*c.ByHeaders, ",")
}

func (c *sectionSingleFlight) GetByQueries() []string {
	if c.ByQueries == nil {
		return []string{}
	}
	return utils.Split(*c.ByQueries, ",")
}

func (c *sectionSingleFlight) GetByBody() bool {
	if c.ByBody == nil {
		return false
	}
	return *c.ByBody
}

func (c *sectionSingleFlight) GetByUserIP() bool {
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

func (c *Configuration) GetLogging() *configLogging {
	logging := c.Logging
	if logging == nil {
		logging = &configLogging{}
	}
	return logging
}

type configLogging struct {
	Enabled *bool `json:"enabled"`
	Format *string `json:"format"`
	Level *string `json:"level"`
	OutputPaths []string `json:"output-paths"`
	ErrorOutputPaths []string `json:"error-output-paths"`
}

func (c *configLogging) GetEnabled() bool {
	if c.Enabled == nil {
		return true
	}
	return *c.Enabled
}

func (c *configLogging) GetFormat() string {
	if c.Format == nil {
		return "flat"
	}
	return *c.Format
}

func (c *configLogging) GetLevel() string {
	if c.Level == nil {
		return "debug"
	}
	return *c.Level
}

func (c *configLogging) GetOutputPaths() []string {
	return c.OutputPaths
}

func (c *configLogging) GetErrorOutputPaths() []string {
	return c.ErrorOutputPaths
}
