package entrypoints

import (
	"fmt"
	"io"
	"net/http"
)

func StartRestServer() (*http.Server) {
	mux := http.NewServeMux()

	mux.HandleFunc("/_/health", makeHealthCheckHandler())
	
	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", 8888),
		MaxHeaderBytes: 1 << 22, // Max header of 4MB
		Handler:        mux,
	}

	s.ListenAndServe()
	
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
