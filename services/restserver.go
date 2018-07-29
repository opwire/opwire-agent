package services

import (
	"fmt"
	"io"
	"net/http"
	"github.com/opwire/opwire-agent/handlers"
)

type ServerOptions struct {
	Host string
	Port int
}

type RestServer struct {
	httpServer *http.Server
	executor *handlers.Executor
}

func NewRestServer(c *ServerOptions) (*RestServer) {
	s := &RestServer{}

	// creates a new command executor
	s.executor, _ = handlers.NewExecutor("ls")

	// defines HTTP request handlers
	mux := http.NewServeMux()

	mux.HandleFunc("/_/health", makeHealthCheckHandler())
	mux.HandleFunc("/run", makeInvocationHandler())

	// determines server's parameters
	host := c.Host
	port := DEFAULT_PORT
	if c.Port > 0 {
		port = c.Port
	}

	// creates a new HTTP server
	s.httpServer = &http.Server{
		Addr:           fmt.Sprintf("%s:%d", host, port),
		MaxHeaderBytes: 1 << 22, // Max header of 4MB
		Handler:        mux,
	}

	// listens and waiting for TERM signal for shutting down
	waitForTermSignal(s.httpServer)

	return s
}

func makeHealthCheckHandler() func(http.ResponseWriter, *http.Request) {
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

func makeInvocationHandler() func(http.ResponseWriter, *http.Request) {
	return func (w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete:
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				io.WriteString(w, fmt.Sprintf(`{"action": "run"}`))
				break
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func waitForTermSignal(s *http.Server) (*http.Server) {
	s.ListenAndServe()
	return s
}

const DEFAULT_PORT int = 17779
