package services

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"github.com/opwire/opwire-agent/invokers"
)

type ServerOptions struct {
	Host string
	Port uint
	CommandString string
	SuppressAutoStart bool
}

type AgentServer struct {
	httpServer *http.Server
	reqSerializer *ReqSerializer
	stateStore *StateStore
	executor *invokers.Executor
	initialized bool
}

func NewAgentServer(c *ServerOptions) (s *AgentServer, err error) {
	// creates a new server instance
	s = &AgentServer{}

	// creates a new command executor
	s.executor, err = invokers.NewExecutor(&invokers.ExecutorOptions{
		Command: invokers.CommandDescriptor{
			CommandString: c.CommandString,
		},
	})

	if err != nil {
		return nil, err
	}

	// creates a StateStore instance
	s.stateStore, err = NewStateStore()

	if err != nil {
		return nil, err
	}

	// creates a ReqSerializer instance
	s.reqSerializer, err = NewReqSerializer()

	if err != nil {
		return nil, err
	}

	// defines HTTP request invokers
	mux := http.NewServeMux()

	mux.HandleFunc("/_/health", s.makeHealthCheckHandler())
	mux.HandleFunc("/run", s.makeInvocationHandler())

	// creates a new HTTP server
	s.httpServer = &http.Server{
		Addr:           buildHttpAddr(c),
		MaxHeaderBytes: 1 << 22, // Max header of 4MB
		Handler:        mux,
	}

	// starts the server by default
	if c == nil || !c.SuppressAutoStart {
		if err = s.Start(); err != nil {
			return nil, err
		}
	}

	// marks this instance has been initialized properly
	s.initialized = true

	return s, nil
}

func (s *AgentServer) Start() (error) {
	// listens and waiting for TERM signal for shutting down
	return s.waitForTermSignal()
}

func (s *AgentServer) Shutdown() (error) {
	return nil
}

func (s *AgentServer) makeHealthCheckHandler() func(http.ResponseWriter, *http.Request) {
	return func (w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, fmt.Sprintf(`{"alive": true}`))
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
				ci, _ := s.buildCommandInvocation(r)
				ib, _ := s.buildCommandStdinBuffer(r)
				var ob bytes.Buffer
				var eb bytes.Buffer
				err := s.executor.Run(ib, ci, &ob, &eb)
				if err == nil {
					w.WriteHeader(http.StatusOK)
					w.Header().Set("Content-Type", "application/text")
					io.WriteString(w, string(ob.Bytes()))
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					w.Header().Set("Content-Type", "application/text")
					io.WriteString(w, string(eb.Bytes()))
				}
			break
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func (s *AgentServer) buildCommandInvocation(r *http.Request) (*invokers.CommandInvocation, error) {
	envs := os.Environ()
	if encoded, err := s.reqSerializer.Encode(r); err == nil {
		envs = append(envs, fmt.Sprintf("OPWIRE_REQUEST=%s", encoded))
	} else {
		return nil, err
	}
	return &invokers.CommandInvocation{
		Envs: envs,
	}, nil
}

func (s *AgentServer) buildCommandStdinBuffer(r *http.Request) (*bytes.Buffer, error) {
	if r.Body == nil {
		return nil, nil
	}

	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(data), nil
}

func (s *AgentServer) waitForTermSignal() (error) {
	s.httpServer.ListenAndServe()
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
