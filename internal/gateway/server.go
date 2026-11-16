package gateway

import (
	"context"
	"net"
	"net/http"
	"time"
)

type Timeouts struct {
	Read       time.Duration
	ReadHeader time.Duration
	Write      time.Duration
	Idle       time.Duration
}

type Server struct {
	httpServer *http.Server
}

func New(listenAddr string, handler http.Handler, timeouts Timeouts) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              listenAddr,
			Handler:           handler,
			ReadTimeout:       timeouts.Read,
			ReadHeaderTimeout: timeouts.ReadHeader,
			WriteTimeout:      timeouts.Write,
			IdleTimeout:       timeouts.Idle,
		},
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) Close() error {
	return s.httpServer.Close()
}

func (s *Server) Serve(listener net.Listener) error {
	return s.httpServer.Serve(listener)
}
