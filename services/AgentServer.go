package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"time"
	"github.com/gorilla/mux"
	"github.com/opwire/opwire-agent/config"
	"github.com/opwire/opwire-agent/invokers"
	"github.com/opwire/opwire-agent/utils"
)

type CommandExecutor interface {
	Register(*invokers.CommandDescriptor, ...string) (error)
	StoreSettings(prefix string, settings map[string]interface{}, format string, resourceName string) (error)
	Run(io.Reader, *invokers.CommandInvocation, io.Writer, io.Writer) (*invokers.ExecutionState, error)
}

type ServerEdition struct {
	Revision string   `json:"revision"`
	Version string    `json:"version"`
}

type ServerOptions struct {
	Host string
	Port uint
	ConfigPath string
	DefaultCommand string
	StaticPath map[string]string
	SuppressAutoStart bool
	Edition ServerEdition
}

type AgentServer struct {
	configManager *config.Manager
	httpServer *http.Server
	httpServeAddr string
	httpServeMux *mux.Router
	reqSerializer *ReqSerializer
	stateStore *StateStore
	executor CommandExecutor
	listeningLock int32
	initialized bool
	edition *ServerEdition
}

func NewAgentServer(c *ServerOptions) (s *AgentServer, err error) {
	if c == nil {
		c = &ServerOptions{}
	}

	// creates a new server instance
	s = &AgentServer{}

	// creates a new command executor
	options := &invokers.ExecutorOptions{}
	if len(c.DefaultCommand) > 0 {
		options.DefaultCommand = &invokers.CommandDescriptor{
			CommandString: c.DefaultCommand,
		}
	}
	s.executor, err = invokers.NewExecutor(options)

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
	s.configManager = config.NewManager(c.Edition.Version, c.ConfigPath)
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

	// register the main resource
	if conf != nil && options.DefaultCommand == nil {
		resourceName := invokers.MAIN_RESOURCE
		resource := conf.Main
		s.importResource(resourceName, resource, conf.Settings, conf.SettingsFormat)
	}

	// register the sub-resources
	if conf != nil && conf.Resources != nil {
		for resourceName, resource := range conf.Resources {
			s.importResource(resourceName, &resource, conf.Settings, conf.SettingsFormat)
		}
	}

	// creates a ReqSerializer instance
	s.reqSerializer, err = NewReqSerializer()

	if err != nil {
		return nil, err
	}

	// defines HTTP request invokers
	baseUrl := EXEC_BASEURL
	if conf != nil && conf.HttpServer != nil && conf.HttpServer.BaseUrl != nil {
		baseUrl = *conf.HttpServer.BaseUrl
	}
	if baseUrl == "/" {
		baseUrl = ""
	}
	s.httpServeMux = mux.NewRouter()
	s.httpServeMux.HandleFunc(CTRL_BASEURL + `health`, s.makeHealthCheckHandler())
	s.httpServeMux.HandleFunc(CTRL_BASEURL + `lock`, s.makeLockServiceHandler(true))
	s.httpServeMux.HandleFunc(CTRL_BASEURL + `unlock`, s.makeLockServiceHandler(false))
	s.httpServeMux.HandleFunc(baseUrl + `/{resourceName:` + config.RESOURCE_NAME_PATTERN + `}`, s.makeInvocationHandler())
	s.httpServeMux.HandleFunc(baseUrl + `/`, s.makeInvocationHandler())
	if len(baseUrl) > 0 {
		s.httpServeMux.HandleFunc(baseUrl, s.makeInvocationHandler())
	}

	urlPaths := utils.SortDesc(utils.Keys(c.StaticPath))
	for _, urlPath := range urlPaths {
		filePath := c.StaticPath[urlPath]
		if utils.IsExists(filePath) {
			log.Printf("Map [%s] -> [%s]", urlPath, filePath)
			s.httpServeMux.PathPrefix(urlPath).Handler(http.StripPrefix(urlPath, http.FileServer(http.Dir(filePath))))
		}
	}

	// creates a new HTTP server
	s.httpServeAddr = buildHttpAddr(c)

	// marks this instance has been initialized properly
	s.initialized = true

	// marks the release manifest
	s.edition = &c.Edition

	// starts the server by default
	if !c.SuppressAutoStart {
		if err = s.Start(); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *AgentServer) Start() (error) {
	if s.httpServer == nil {
		s.httpServer = &http.Server{
			Addr:           s.httpServeAddr,
			MaxHeaderBytes: 1 << 22, // Max header of 4MB
			Handler:        s.httpServeMux,
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

func (s *AgentServer) importResource(resourceName string, resource *invokers.CommandEntrypoint,
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
		switch r.Method {
		case
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete:
				ib, ibErr := s.buildCommandStdinBuffer(r)
				if ibErr != nil {
					w.Header().Set("X-Error-Message", ibErr.Error())
					w.WriteHeader(http.StatusBadRequest)
					break
				}
				if ib != nil {
					defer ib.Close()
				}
				ci, ciErr := s.buildCommandInvocation(r)
				if ciErr != nil {
					w.Header().Set("X-Error-Message", ciErr.Error())
					w.WriteHeader(http.StatusBadRequest)
					break
				}
				var ob bytes.Buffer
				var eb bytes.Buffer
				state, err := s.executor.Run(ib, ci, &ob, &eb)
				if state != nil && state.IsTimeout {
					w.Header().Set("Content-Type","text/plain")
					w.WriteHeader(http.StatusRequestTimeout)
					io.WriteString(w, "Running processes are killed")
					break
				}
				if err != nil {
					w.Header().Set("X-Error-Message", err.Error())
					w.Header().Set("Content-Type","text/plain")
					w.WriteHeader(http.StatusInternalServerError)
					io.WriteString(w, string(eb.Bytes()))
					break
				}
				w.Header().Set("X-Exec-Duration", fmt.Sprintf("%f", state.Duration.Seconds()))
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)
				io.WriteString(w, string(ob.Bytes()))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

var ACCEPTED_METHODS []string = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
}

func normalizeMethod(method string) (string, bool) {
	name := strings.ToUpper(method)
	if utils.Contains(ACCEPTED_METHODS, name) {
		return name, true
	}
	return name, false
}

func (s *AgentServer) buildCommandInvocation(r *http.Request) (*invokers.CommandInvocation, error) {
	// extract command identifier
	params := mux.Vars(r)
	resourceName := params["resourceName"]
	methodName := r.Method
	log.Printf("Command [%s#%s] has been invoked", resourceName, methodName)
	// prepare environment variables
	envs := os.Environ()
	// import the release information
	if s.edition != nil {
		if str, err := json.Marshal(s.edition); err == nil {
			envs = append(envs, fmt.Sprintf("%s=%s", OPWIRE_EDITION_PREFIX, str))
		}
	}
	// build the request query & data
	if encoded, err := s.reqSerializer.Encode(r); err == nil {
		envs = append(envs, fmt.Sprintf("%s=%s", OPWIRE_REQUEST_PREFIX, encoded))
	} else {
		return nil, err
	}
	// return the CommandInvocation reference
	return &invokers.CommandInvocation{
		Envs: envs,
		ResourceName: resourceName,
		MethodName: methodName,
	}, nil
}

func (s *AgentServer) buildCommandStdinBuffer(r *http.Request) (io.ReadCloser, error) {
	return r.Body, nil
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

func buildHttpAddr(c *ServerOptions) string {
	port := DEFAULT_PORT
	if c.Port > 0 {
		port = c.Port
	}
	return fmt.Sprintf("%s:%d", c.Host, port)
}

const CTRL_BASEURL string = `/_/`
const EXEC_BASEURL string = `/run`
const DEFAULT_PORT uint = 17779
const OPWIRE_EDITION_PREFIX string = "OPWIRE_EDITION"
const OPWIRE_REQUEST_PREFIX string = "OPWIRE_REQUEST"
const OPWIRE_SETTINGS_PREFIX string = "OPWIRE_SETTINGS"
