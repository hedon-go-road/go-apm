package dogapm

import (
	"context"
	"log"
	"net/http"
	"time"
)

type HTTPServer struct {
	mux *http.ServeMux
	*http.Server
}

func NewHTTPServer(addr string) *HTTPServer {
	mux := http.NewServeMux()
	s := &HTTPServer{
		mux: mux,
		Server: &http.Server{
			Addr:              addr,
			Handler:           mux,
			ReadHeaderTimeout: 30 * time.Second, //nolint:mnd
		},
	}
	globalStarters = append(globalStarters, s)
	globalClosers = append(globalClosers, s)
	return s
}

func (s *HTTPServer) Close() {
	if s.Server != nil {
		if err := s.Server.Shutdown(context.Background()); err != nil {
			log.Println("Error stopping server:", err)
		}
	}
}

func (s *HTTPServer) Handle(pattern string, handler http.Handler) {
	s.mux.Handle(pattern, handler)
}

func (s *HTTPServer) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.mux.HandleFunc(pattern, handler)
}

func (s *HTTPServer) Start() {
	go func() {
		log.Println("statring http server")
		if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Error starting server:", err)
		}
	}()
}

func (s *HTTPServer) Stop() {
	if err := s.Server.Shutdown(context.Background()); err != nil {
		log.Fatal("Error stopping server:", err)
	}
}
