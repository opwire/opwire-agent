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
	Run(*bytes.Buffer, *invokers.CommandInvocation, *bytes.Buffer, *bytes.Buffer) (*invokers.ExecutionState, error)
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
	configLoader *config.Loader
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
	s.configLoader = config.NewLoader(c.Edition.Version, c.ConfigPath)
	conf, result, err := s.configLoader.Load()
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
		resource := conf.Main
		if resource != nil {
			s.executor.Register(resource.Default, invokers.MAIN_RESOURCE)
			if len(resource.Methods) > 0 {
				for methodName, methodDescriptor := range resource.Methods {
					if methodId, ok := normalizeMethod(methodName); ok {
						s.executor.Register(methodDescriptor, invokers.MAIN_RESOURCE, methodId)
					}
				}
			}
		}
	}

	// register the commands
	if conf != nil && conf.Resources != nil {
		for resourceId, resource := range conf.Resources {
			s.executor.Register(resource.Default, resourceId)
			if len(resource.Methods) > 0 {
				for methodName, methodDescriptor := range resource.Methods {
					if methodId, ok := normalizeMethod(methodName); ok {
						s.executor.Register(methodDescriptor, resourceId, methodId)
					}
				}
			}
		}
	}

	// creates a ReqSerializer instance
	s.reqSerializer, err = NewReqSerializer()

	if err != nil {
		return nil, err
	}

	// defines HTTP request invokers
	s.httpServeMux = mux.NewRouter()
	s.httpServeMux.HandleFunc("/_/health", s.makeHealthCheckHandler())
	s.httpServeMux.HandleFunc("/run/{resourceName:[a-zA-Z0-9_-]*}", s.makeInvocationHandler())
	s.httpServeMux.HandleFunc("/run", s.makeInvocationHandler())

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

func (s *AgentServer) makeHealthCheckHandler() func(http.ResponseWriter, *http.Request) {
	return func (w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if s.isReady() {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				io.WriteString(w, fmt.Sprintf(`{"alive": true}`))
				break
			}
			w.WriteHeader(http.StatusServiceUnavailable)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func (s *AgentServer) makeInvocationHandler() func(http.ResponseWriter, *http.Request) {
	return func (w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete:
				ci, ciErr := s.buildCommandInvocation(r)
				if ciErr != nil {
					w.Header().Set("X-Error-Message", ciErr.Error())
					w.WriteHeader(http.StatusBadRequest)
					break
				}
				ib, ibErr := s.buildCommandStdinBuffer(r)
				if ibErr != nil {
					w.Header().Set("X-Error-Message", ibErr.Error())
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
	// prepare environment variables
	envs := os.Environ()
	// import the release information
	if s.edition != nil {
		if str, err := json.Marshal(s.edition); err == nil {
			envs = append(envs, fmt.Sprintf("OPWIRE_EDITION=%s", str))
		}
	}
	// build the request query & data
	if encoded, err := s.reqSerializer.Encode(r); err == nil {
		envs = append(envs, fmt.Sprintf("OPWIRE_REQUEST=%s", encoded))
	} else {
		return nil, err
	}
	// extract command identifier
	params := mux.Vars(r)
	resourceId := params["resourceName"]
	log.Printf("Command [%s] has been invoked", resourceId)
	return &invokers.CommandInvocation{
		Envs: envs,
		ResourceName: resourceId,
		MethodName: r.Method,
	}, nil
}

func (s *AgentServer) buildCommandStdinBuffer(r *http.Request) (*bytes.Buffer, error) {
	if r.Body == nil {
		return bytes.NewBuffer(nil), nil
	}

	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(data), nil
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

const DEFAULT_PORT uint = 17779
