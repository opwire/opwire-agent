package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
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
	listeningLock int32
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

	// marks this instance has been initialized properly
	s.initialized = true

	// starts the server by default
	if c == nil || !c.SuppressAutoStart {
		if err = s.Start(); err != nil {
			return nil, err
		}
	}

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
			if s.isReady() {
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				io.WriteString(w, fmt.Sprintf(`{"alive": true}`))
				break
			}
			w.WriteHeader(http.StatusServiceUnavailable)
			break
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			break
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
				if err := s.executor.Run(ib, ci, &ob, &eb); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Header().Set("Content-Type","text/plain")
					io.WriteString(w, string(eb.Bytes()))
					break
				}
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "text/plain")
				io.WriteString(w, string(ob.Bytes()))
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
	closingTimeout := time.Second*5
	idleConnections := make(chan struct{})

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGTERM, syscall.SIGTSTP)
		<-sig

		s.lockService()

		fmt.Println()

		log.Printf("SIGTERM/SIGTSTP received, no new requests allowed. wait for %s\n", closingTimeout.String())
		<-time.Tick(closingTimeout)

		if err := s.httpServer.Shutdown(context.Background()); err != nil {
			log.Printf("httpServer.Shutdown() failed: %v", err)
		}

		log.Printf("HTTP server is shutting down in: %s\n", closingTimeout.String())
		<-time.Tick(closingTimeout)

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
