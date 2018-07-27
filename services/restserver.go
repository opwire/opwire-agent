package services

import (
	"fmt"
	"io"
	"net/http"
)

type ServerConfig struct {
	Host string
	Port int
}

func StartRestServer(c *ServerConfig) (*http.Server) {
	mux := http.NewServeMux()

	mux.HandleFunc("/_/health", makeHealthCheckHandler())

	host := c.Host
	port := 17779
	if c.Port > 0 {
    port = c.Port
	}

	s := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", host, port),
		MaxHeaderBytes: 1 << 22, // Max header of 4MB
		Handler:        mux,
	}

	return waitForTermSignal(s)
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

func waitForTermSignal(s *http.Server) (*http.Server) {
	s.ListenAndServe()
	return s
}
