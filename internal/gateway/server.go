package gateway

import (
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func New(listenAddr string, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              listenAddr,
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}
