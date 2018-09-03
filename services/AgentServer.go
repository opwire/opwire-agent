package services

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"github.com/opwire/opwire-agent/handlers"
)

type ServerOptions struct {
	Host string
	Port uint
	CommandString string
	SuppressAutoStart bool
}

type AgentServer struct {
	httpServer *http.Server
	stateStore *StateStore
	executor *handlers.Executor
	initialized bool
}

func NewAgentServer(c *ServerOptions) (*AgentServer) {
	// creates a new server instance
	s := &AgentServer{}

	// creates a new command executor
	s.executor, _ = handlers.NewExecutor(&handlers.ExecutorOptions{
		Command: handlers.CommandDescriptor{
			CommandString: c.CommandString,
		},
	})

	// defines HTTP request handlers
	mux := http.NewServeMux()

	mux.HandleFunc("/_/health", s.makeHealthCheckHandler())
	mux.HandleFunc("/run", s.makeInvocationHandler())

	// determines server's parameters
	host := c.Host
	port := DEFAULT_PORT
	if c.Port > 0 {
		port = c.Port
	}

	// creates a StateStore instance
	s.stateStore = NewStateStore()

	// creates a new HTTP server
	s.httpServer = &http.Server{
		Addr:           fmt.Sprintf("%s:%d", host, port),
		MaxHeaderBytes: 1 << 22, // Max header of 4MB
		Handler:        mux,
	}

	// marks this instance has been initialized properly
	s.initialized = true

	if c == nil || !c.SuppressAutoStart {
		s.Start()
	}
	return s
}

func (s *AgentServer) Start() (*AgentServer, error) {
	// listens and waiting for TERM signal for shutting down
	waitForTermSignal(s.httpServer)
	return s, nil
}

func (s *AgentServer) Shutdown() (*AgentServer, error) {
	return s, nil
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
				pi, _ := s.buildCommandStdinData(r)
				outData, errData, err := s.executor.Run(ci, pi)
				if err == nil {
					w.WriteHeader(http.StatusOK)
					w.Header().Set("Content-Type", "application/text")
					io.WriteString(w, string(outData))
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					w.Header().Set("Content-Type", "application/text")
					io.WriteString(w, string(errData))
				}
			break
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func (s *AgentServer) buildCommandInvocation(r *http.Request) (*handlers.CommandInvocation, error) {
	return &handlers.CommandInvocation{}, nil
}

func (s *AgentServer) buildCommandStdinData(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}

	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func waitForTermSignal(s *http.Server) (*http.Server) {
	s.ListenAndServe()
	return s
}

const DEFAULT_PORT uint = 17779
