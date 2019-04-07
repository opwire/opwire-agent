package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync/atomic"
	"time"
	"github.com/gorilla/mux"
	"github.com/opwire/opwire-agent/config"
	"github.com/opwire/opwire-agent/invokers"
	"github.com/opwire/opwire-agent/utils"
)

type CommandExecutor interface {
	Register(descriptor *invokers.CommandDescriptor, names ...string) (error)
	ResolveCommandDescriptor(opts *invokers.CommandInvocation) (descriptor *invokers.CommandDescriptor, resourceName *string, methodName *string, err error)
	GetSettings(resourceName string) []string
	StoreSettings(prefix string, settings map[string]interface{}, format string, resourceName string) (error)
	Run(io.Reader, *invokers.CommandInvocation, io.Writer, io.Writer) (*invokers.ExecutionState, error)
}

type AgentServerOptions interface {
	GetConfigPath() string
	GetDirectCommand() string
	GetHost() string
	GetPort() uint
	GetStaticPath() map[string]string
	SuppressAutoStart() bool
	GetRevision() string
	GetVersion() string
}

type AgentServer struct {
	configManager *config.Manager
	httpServer *http.Server
	httpRouter *mux.Router
	httpOptions *httpServerOptions
	reqRestrictor *ReqRestrictor
	reqSerializer *ReqSerializer
	stateStore *StateStore
	executor CommandExecutor
	options AgentServerOptions
	listeningLock int32
	explanationEnabled bool
}

type httpServerOptions struct {
	Addr string
	MaxHeaderBytes int
}

func NewAgentServer(o AgentServerOptions) (s *AgentServer, err error) {
	if o == nil {
		return nil, fmt.Errorf("AgentServerOptions must not be nil")
	}

	// creates a new server instance
	s = &AgentServer{}

	// remember server edition & options
	s.options = o

	// creates a new command executor
	s.executor, err = invokers.NewExecutor(&invokers.ExecutorOptions{})

	if err != nil {
		return nil, err
	}

	// creates a StateStore instance
	s.stateStore, err = NewStateStore()

	if err != nil {
		return nil, err
	}

	// load configuration
	var conf *config.Configuration

	// determine configuration path
	s.configManager = config.NewManager(s.options)
	conf, result, err := s.configManager.Load()
	if err != nil {
		return nil, err
	}
	if result != nil && !result.Valid() {
		errstrs := []string {"The configuration is not valid. Errors:"}
		for _, desc := range result.Errors() {
			errstrs = append(errstrs, fmt.Sprintf("%s", desc))
		}
		return nil, fmt.Errorf(strings.Join(errstrs, "\n - "))
	}
	if conf == nil {
		conf = &config.Configuration{}
	}

	// register main & sub-resources
	s.registerResources(conf)

	// creates a ReqSerializer instance
	s.reqSerializer, err = NewReqSerializer()

	if err != nil {
		return nil, err
	}

	// create a weighted semaphore
	var reqRestrictorOpts ReqRestrictorOptions
	if conf.HttpServer != nil {
		reqRestrictorOpts = conf.HttpServer
	}
	s.reqRestrictor, err = NewReqRestrictor(reqRestrictorOpts)

	if err != nil {
		return nil, err
	}

	// defines HTTP request invokers
	s.httpRouter = mux.NewRouter()
	s.httpRouter.HandleFunc(CTRL_BASEURL + `/health`, s.makeHealthCheckHandler())
	s.httpRouter.HandleFunc(CTRL_BASEURL + `/lock`, s.makeLockServiceHandler(true))
	s.httpRouter.HandleFunc(CTRL_BASEURL + `/unlock`, s.makeLockServiceHandler(false))

	s.mappingResourceToExecUrl(EXEC_BASEURL, conf)
	s.mappingResourceToExecUrl(EXEC_BASEURL_DEPRECATED, conf)

	// validate duplicated resource patterns
	if err := validateResourcePatterns(conf); err != nil {
		return nil, err
	}

	// declare resource patterns
	s.mappingResourcePatterns(conf)

	webStaticPath := s.options.GetStaticPath()
	urlPaths := utils.SortDesc(utils.Keys(webStaticPath))
	for _, urlPath := range urlPaths {
		filePath := webStaticPath[urlPath]
		if utils.IsExists(filePath) {
			log.Printf("Map [%s] -> [%s]", urlPath, filePath)
			s.httpRouter.PathPrefix(urlPath).Handler(http.StripPrefix(urlPath, http.FileServer(http.Dir(filePath))))
		}
	}

	// creates a new HTTP server
	s.httpOptions = new(httpServerOptions)
	s.httpOptions.Addr = buildHttpAddr(s.options, conf)
	s.httpOptions.MaxHeaderBytes = 1 << 22 // new default: 4MB

	// other configurations
	if conf.Agent != nil && conf.Agent.ExplanationEnabled != nil {
		s.explanationEnabled = *conf.Agent.ExplanationEnabled
	}

	// starts the server by default
	if !s.options.SuppressAutoStart() {
		if err = s.Start(); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *AgentServer) Start() (error) {
	if s.httpServer == nil {
		s.httpServer = &http.Server{
			Addr: s.httpOptions.Addr,
			MaxHeaderBytes: s.httpOptions.MaxHeaderBytes,
			Handler: s.httpRouter,
		}
	}
	// listens and waiting for TERM signal for shutting down
	return s.listenAndServe()
}

func (s *AgentServer) Shutdown() (error) {
	closingTimeout := time.Second * 5

	defer func() {
		s.httpServer = nil
	}()

	if s.isReady() {
		if err := s.lockService(); err != nil {
			log.Printf("lockService() failed: %v", err)
		}

		log.Printf("No new requests allowed, wait for %s\n", closingTimeout.String())
		<-time.Tick(closingTimeout)
	}

	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(context.Background()); err != nil {
			log.Printf("httpServer.Shutdown() failed: %v", err)
		}

		log.Printf("HTTP server is shutting down in: %s\n", closingTimeout.String())
		<-time.Tick(closingTimeout)
	}

	return nil
}

func (s *AgentServer) registerResources(conf *config.Configuration) {
	// register the main resource
	if conf.Main != nil {
		resourceName := invokers.MAIN_RESOURCE
		resource := conf.Main
		s.registerResource(resourceName, resource, conf.Settings, conf.SettingsFormat)
	}

	// register the sub-resources
	if conf.Resources != nil {
		for resourceName, resource := range conf.Resources {
			s.registerResource(resourceName, &resource, conf.Settings, conf.SettingsFormat)
		}
	}
}

func (s *AgentServer) registerResource(resourceName string, resource *invokers.CommandEntrypoint,
		settings map[string]interface{}, format *string) {
	if resource != nil {
		s.executor.Register(resource.Default, resourceName)
		if len(resource.Methods) > 0 {
			for methodName, methodDescriptor := range resource.Methods {
				if methodId, ok := normalizeMethod(methodName); ok {
					s.executor.Register(methodDescriptor, resourceName, methodId)
				}
			}
		}
		if privSettings, err := utils.CombineSettings(resource.Settings, settings); err == nil {
			privFormat := "json"
			if format != nil {
				privFormat = *format
			}
			if resource.SettingsFormat != nil {
				privFormat = *resource.SettingsFormat
			}
			s.executor.StoreSettings(OPWIRE_SETTINGS_PREFIX, privSettings, privFormat, resourceName)
		}
	}
}

func (s *AgentServer) mappingResourcePatterns(conf *config.Configuration) {
	// register the main resource
	if conf.Main != nil {
		resourceName := invokers.MAIN_RESOURCE
		resourceConf := conf.Main
		s.mappingResourcePattern(resourceName, resourceConf)
	}

	// register the sub-resources
	if conf.Resources != nil {
		for resourceName, resourceConf := range conf.Resources {
			s.mappingResourcePattern(resourceName, &resourceConf)
		}
	}
}

func (s *AgentServer) mappingResourcePattern(resourceName string, resourceConf *invokers.CommandEntrypoint) {
	if len(resourceName) > 0 && resourceConf.Pattern != nil {
		s.httpRouter.HandleFunc(*resourceConf.Pattern, s.makeUrlPatternHandler(resourceName))
	}
}

func (s *AgentServer) mappingResourceToExecUrl(defaultBaseUrl string, conf *config.Configuration) {
	baseUrl := buildExecUrl(defaultBaseUrl, conf)
	s.httpRouter.HandleFunc(baseUrl + `/{resourceName:` + config.RESOURCE_NAME_PATTERN + `}`, s.makeInvocationHandler())
	s.httpRouter.HandleFunc(baseUrl + `/`, s.makeInvocationHandler())
	if len(baseUrl) > 0 {
		s.httpRouter.HandleFunc(baseUrl, s.makeInvocationHandler())
	}
}

func (s *AgentServer) makeLockServiceHandler(freezed bool) func(http.ResponseWriter, *http.Request) {
	return func (w http.ResponseWriter, r *http.Request) {
		if freezed {
			s.lockService()
		} else {
			s.unlockService()
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (s *AgentServer) makeHealthCheckHandler() func(http.ResponseWriter, *http.Request) {
	return func (w http.ResponseWriter, r *http.Request) {
		if !s.isReady() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			io.WriteString(w, `{"ready": false}`)
			return
		}
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{"ready": true, "alive": true}`)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func (s *AgentServer) makeInvocationHandler() func(http.ResponseWriter, *http.Request) {
	return func (w http.ResponseWriter, r *http.Request) {
		if !s.isReady() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		if !isMethodAccepted(r.Method) {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		params := mux.Vars(r)
		resourceName := params["resourceName"]
		s.doExecuteCommand(w, r, resourceName, true)
	}
}

func (s *AgentServer) makeUrlPatternHandler(resourceName string) func(http.ResponseWriter, *http.Request) {
	return func (w http.ResponseWriter, r *http.Request) {
		if !s.isReady() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		if !isMethodAccepted(r.Method) {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		s.doExecuteCommand(w, r, resourceName, false)
	}
}

func (s *AgentServer) doExecuteCommand(w http.ResponseWriter, r *http.Request, resourceName string, defaultUrl bool) {
	expIn, expOut, expErr := s.getExplanationModes(r)

	ib, tee := s.generateTeeBuffer()

	ir, irErr := s.buildCommandStdinReader(r, tee)
	if irErr != nil {
		w.Header().Set(RES_HEADER_ERROR_MESSAGE, irErr.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if ir != nil {
		defer ir.Close()
	}
	ci, ciErr := s.buildCommandInvocation(r, resourceName, defaultUrl)
	if ciErr != nil {
		w.Header().Set(RES_HEADER_ERROR_MESSAGE, ciErr.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if expIn {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusResetContent)
		ioutil.ReadAll(ir)
		s.explainRequest(w, ib, ci)
		return
	}
	var ob bytes.Buffer
	var eb bytes.Buffer
	var state *invokers.ExecutionState
	var err error

	if s.reqRestrictor.HasSemaphore() {
		if err := s.reqRestrictor.Acquire(1); err != nil {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, fmt.Sprintf("Failed to acquire permits, error: [%v]", err))
			return
		}
		defer s.reqRestrictor.Release(1)
	}

	if s.reqRestrictor.HasSingleFlight() {
		_state, _err, _ := s.reqRestrictor.FilterByDigest(r, func() (interface{}, error) {
			return s.executor.Run(ir, ci, &ob, &eb)
		})
		state = _state.(*invokers.ExecutionState)
		err = _err
	} else {
		state, err = s.executor.Run(ir, ci, &ob, &eb)
	}

	if state != nil && state.IsTimeout {
		w.Header().Set("Content-Type", "text/plain")
		writeHeaderExecDuration(w, state)
		w.WriteHeader(http.StatusRequestTimeout)
		io.WriteString(w, "Running processes are killed")
		return
	}
	if err != nil {
		if expErr {
			w.Header().Set("Content-Type", "text/plain")
			writeHeaderExecDuration(w, state)
			w.WriteHeader(http.StatusInternalServerError)
			s.explainResult(w, ib, ci, err, &ob, &eb)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		writeHeaderExecDuration(w, state)
		w.Header().Set(RES_HEADER_ERROR_MESSAGE, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, string(eb.Bytes()))
		return
	} else {
		if expOut {
			w.Header().Set("Content-Type", "text/plain")
			writeHeaderExecDuration(w, state)
			w.WriteHeader(http.StatusResetContent)
			s.explainResult(w, ib, ci, err, &ob, &eb)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		writeHeaderExecDuration(w, state)
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, string(ob.Bytes()))
		return
	}
}

func writeHeaderExecDuration(w http.ResponseWriter, state *invokers.ExecutionState) {
	if state == nil || state.Duration == 0 {
		return
	}
	w.Header().Set(RES_HEADER_EXEC_DURATION, fmt.Sprintf("%f", state.Duration.Seconds()))
}

func (s *AgentServer) generateTeeBuffer() (*bytes.Buffer, io.Writer) {
	var ib bytes.Buffer
	var tee io.Writer
	if s.explanationEnabled {
		tee = &ib
	}
	return &ib, tee
}

func (s *AgentServer) getExplanationModes(r *http.Request) (bool, bool, bool) {
	if !s.explanationEnabled {
		return false, false, false
	}
	return len(r.Header.Get(REQ_HEADER_SUPPRESS_EXECUTION)) > 0,
		len(r.Header.Get(REQ_HEADER_EXPLAIN_SUCCESS)) > 0,
		len(r.Header.Get(REQ_HEADER_EXPLAIN_FAILURE)) > 0
}

func isMethodAccepted(method string) (bool) {
	switch method {
	case
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete:
		return true
	default:
		return false
	}
}

func normalizeMethod(method string) (string, bool) {
	name := strings.ToUpper(method)
	if isMethodAccepted(name) {
		return name, true
	}
	return name, false
}

func (s *AgentServer) buildCommandInvocation(r *http.Request, resourceName string, defaultUrl bool) (*invokers.CommandInvocation, error) {
	// extract command identifier
	methodName := r.Method
	log.Printf("Command [%s#%s] has been invoked", resourceName, methodName)
	// prepare environment variables
	envs := os.Environ()
	// import the release information
	if s.options != nil {
		edition := map[string]string {
			"revision": s.options.GetRevision(),
			"version": s.options.GetVersion(),
		}
		if str, err := json.Marshal(edition); err == nil {
			envs = append(envs, fmt.Sprintf("%s=%s", OPWIRE_EDITION_PREFIX, str))
		}
	}
	// build the request query & data
	if encoded, err := s.reqSerializer.Encode(r, defaultUrl); err == nil {
		envs = append(envs, fmt.Sprintf("%s=%s", OPWIRE_REQUEST_PREFIX, encoded))
	} else {
		return nil, err
	}
	// create a new CommandInvocation
	ci := &invokers.CommandInvocation{
		Envs: envs,
		ResourceName: resourceName,
		MethodName: methodName,
	}
	// determine the direct-command
	if s.options != nil && len(s.options.GetDirectCommand()) > 0 {
		ci.DirectCommand = s.options.GetDirectCommand()
	}
	// determine customized execution timeout
	timeout, err := time.ParseDuration(r.Header.Get(REQ_HEADER_EXECUTION_TIMEOUT))
	if err == nil && timeout > 0 {
		ci.ExecutionTimeout = invokers.ConvertDurationToSecond(timeout)
	}
	// create the Context
	ci.Context = r.Context()
	// return the CommandInvocation reference
	return ci, nil
}

type teeReadCloser struct {
	io.Reader
	io.Closer
}

func (s *AgentServer) buildCommandStdinReader(r *http.Request, w io.Writer) (io.ReadCloser, error) {
	src := r.Body
	if src == nil {
		return nil, nil
	}
	if w != nil {
		src = &teeReadCloser{io.TeeReader(src, w), src}
	}
	return src, nil
}

func (s *AgentServer) explainRequest(w http.ResponseWriter, ib *bytes.Buffer, ci *invokers.CommandInvocation) {
	// display agent's release information
	edition, p1 := utils.FirstHasPrefix(ci.Envs, OPWIRE_EDITION_PREFIX_PLUS, true)
	if p1 >= 0 {
		printJsonString(w, "edition", edition)
	}

	// display the request parameters
	reqText, p2 := utils.FirstHasPrefix(ci.Envs, OPWIRE_REQUEST_PREFIX_PLUS, true)
	if p2 >= 0 {
		printJsonString(w, "request", reqText)
	}

	// display the resource, method and command
	var resourceRef, methodRef *string
	if len(ci.DirectCommand) == 0 {
		commandInfo := make(map[string]interface{}, 0)

		providedInfo := map[string]interface{}{
			"resource": ci.ResourceName,
			"method": ci.MethodName,
		}
		if ci.ExecutionTimeout > 0 {
			providedInfo["timeout"] = ci.ExecutionTimeout
		}
		commandInfo["provided"] = providedInfo

		var descriptor *invokers.CommandDescriptor
		descriptor, resourceRef, methodRef, _ = s.executor.ResolveCommandDescriptor(ci)
		resolvedInfo := map[string]interface{}{
			"resource": resourceRef,
			"method": methodRef,
		}
		if descriptor != nil {
			resolvedInfo["command"] = descriptor.CommandString
			timeout := invokers.GetExecutionTimeout(descriptor, ci)
			if timeout > 0 {
				resolvedInfo["timeout"] = timeout
			}
		}
		commandInfo["resolved"] = resolvedInfo

		printJsonObject(w, "command", commandInfo)
	} else {
		printSection(w, "command", fmt.Sprintf(`direct-command: "%s"`, ci.DirectCommand))
	}

	// display the settings
	if resourceRef != nil {
		settingsEnvs := s.executor.GetSettings(*resourceRef)
		settingsText, p3 := utils.FirstHasPrefix(settingsEnvs, OPWIRE_SETTINGS_PREFIX_PLUS, true)
		if p3 >= 0 {
			printJsonString(w, "settings", settingsText)
		} else {
			printCollection(w, "settings", settingsEnvs)
		}
	}

	// display the input from stdin
	printSection(w, "stdin", ib.Bytes())
}

func (s *AgentServer) explainResult(w http.ResponseWriter, ib *bytes.Buffer, ci *invokers.CommandInvocation, err error, ob *bytes.Buffer, eb *bytes.Buffer) {
	s.explainRequest(w, ib, ci)

	if err != nil {
		printSection(w, "stderr", eb.Bytes())
		printSection(w, "error", []byte(err.Error()))
	} else {
		printSection(w, "stdout", ob.Bytes())
	}
}

func printSection(w http.ResponseWriter, label string, data interface{}) {
	header := utils.PadString("[" + label, utils.LEFT, 80, "-")
	footer := utils.PadString(label + "]", utils.RIGHT, 80, "-")
	io.WriteString(w, fmt.Sprintf("\n%s\n%s\n%s\n", header, data, footer))
}

func printCollection(w http.ResponseWriter, label string, settings []string) {
	if len(settings) > 0 {
		lines := utils.Map(settings, func(s string, i int) string {
			return fmt.Sprintf("%d) %s", (i + 1), s)
		})
		section := strings.Join(lines, "\n")
		printSection(w, label, section)
	}
}

func printJsonString(w http.ResponseWriter, label string, dataText string) error {
	if len(dataText) > 0 {
		dataMap := make(map[string]interface{}, 0)
		err := json.Unmarshal([]byte(dataText), &dataMap)
		if err == nil {
			return printJsonObject(w, label, dataMap)
		} else {
			printSection(w, label + " (text)", dataText)
		}
	}
	return nil
}

func printJsonObject(w http.ResponseWriter, label string, dataMap map[string]interface{}) error {
	if len(dataMap) > 0 {
		dataJson, err := json.MarshalIndent(dataMap, "", "  ")
		if err != nil {
			return err
		}
		if len(dataJson) == 0 {
			return fmt.Errorf("Marshalling failed, data is empty")
		}
		printSection(w, label, dataJson)
	}
	return nil
}

func (s *AgentServer) listenAndServe() (error) {
	idleConnections := make(chan struct{})

	go func() {
		SIGLIST := utils.ShutdownSignals()
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, SIGLIST...)
		<-sig

		fmt.Println()
		log.Printf("SIGTERM/SIGTSTP received. Agent is shutting down ...\n")

		s.Shutdown()

		close(idleConnections)
	}()

	go func() {
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("httpServer.ListenAndServe() failed: %v", err)
			close(idleConnections)
		}
	}()

	s.unlockService()

	<-idleConnections
	return nil
}

func (s *AgentServer) isReady() bool {
	return atomic.LoadInt32(&s.listeningLock) != 0
}

func (s *AgentServer) lockService() (error) {
	atomic.StoreInt32(&s.listeningLock, 0)
	return nil
}

func (s *AgentServer) unlockService() (error) {
	atomic.StoreInt32(&s.listeningLock, 1)
	return nil
}

func buildHttpAddr(opts AgentServerOptions, c *config.Configuration) string {
	var conf *config.ConfigHttpServer
	if c != nil {
		conf = c.HttpServer
	}
	var host string
	if conf != nil && conf.Host != nil {
		host = *conf.Host
	}
	if opts != nil && opts.GetHost() != "" {
		host = opts.GetHost()
	}

	port := DEFAULT_PORT
	if conf != nil && conf.Port != nil {
		port = *conf.Port
	}
	if opts != nil && opts.GetPort() != 0 {
		port = opts.GetPort()
	}

	return fmt.Sprintf("%s:%d", host, port)
}

func buildExecUrl(defaultBaseUrl string, conf *config.Configuration) string {
	baseUrl := defaultBaseUrl
	if conf != nil && conf.HttpServer != nil && conf.HttpServer.BaseUrl != nil {
		baseUrl = *conf.HttpServer.BaseUrl
	}
	if baseUrl == "/" {
		baseUrl = ""
	}
	return baseUrl
}

func validateResourcePatterns(conf *config.Configuration) error {
	patterns := make(map[string][]string, 0)

	if conf.Main != nil {
		resourceConf := conf.Main
		countDuplicatedPatterns(patterns, resourceConf)
	}

	if conf.Resources != nil {
		for _, resourceConf := range conf.Resources {
			countDuplicatedPatterns(patterns, &resourceConf)
		}
	}

	errstrs := []string {}
	for _, dup := range patterns {
		if len(dup) > 1 {
			errstrs = append(errstrs, strings.Join(dup, ", "))
		}
	}
	if len(errstrs) > 0 {
		errstrs = append([]string {"Command url patterns are duplicated. Errors:"}, errstrs...)
		return fmt.Errorf(strings.Join(errstrs, "\n - "))
	}

	return nil
}

var re *regexp.Regexp = regexp.MustCompile(`{([^{]*)}`)

func countDuplicatedPatterns(patterns map[string][]string, resourceConf *invokers.CommandEntrypoint) {
	if resourceConf.Pattern != nil {
		s := re.ReplaceAllString(*resourceConf.Pattern, "*")
		if _, ok := patterns[s]; !ok {
			patterns[s] = make([]string, 0)
		}
		patterns[s] = append(patterns[s], *resourceConf.Pattern)
	}
}

const CTRL_BASEURL string = `/_`
const EXEC_BASEURL string = `/$`
const EXEC_BASEURL_DEPRECATED string = `/run`
const DEFAULT_PORT uint = 17779
const OPWIRE_EDITION_PREFIX string = "OPWIRE_EDITION"
const OPWIRE_EDITION_PREFIX_PLUS string = OPWIRE_EDITION_PREFIX + "="
const OPWIRE_REQUEST_PREFIX string = "OPWIRE_REQUEST"
const OPWIRE_REQUEST_PREFIX_PLUS string = OPWIRE_REQUEST_PREFIX + "="
const OPWIRE_SETTINGS_PREFIX string = "OPWIRE_SETTINGS"
const OPWIRE_SETTINGS_PREFIX_PLUS string = OPWIRE_SETTINGS_PREFIX + "="

const REQ_HEADER_REQUEST_ID_NAME string = "Opwire-Request-Id"
const REQ_HEADER_EXECUTION_TIMEOUT string = "Opwire-Execution-Timeout"
const REQ_HEADER_SUPPRESS_EXECUTION string = "Opwire-Suppress-Running"
const REQ_HEADER_EXPLAIN_SUCCESS string = "Opwire-Explain-Success"
const REQ_HEADER_EXPLAIN_FAILURE string = "Opwire-Explain-Failure"

const RES_HEADER_ERROR_MESSAGE string = "X-Error-Message"
const RES_HEADER_EXEC_DURATION string = "X-Exec-Duration"
