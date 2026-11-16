package main

import (
	"log/slog"
	"net/http"
	"sync/atomic"

	"github.com/HitanshuDev/aegis-api-gateway/internal/config"
	"github.com/HitanshuDev/aegis-api-gateway/internal/middleware"
	"github.com/HitanshuDev/aegis-api-gateway/internal/proxy"
)

type healthState struct {
	ready atomic.Bool
}

func newHealthState() *healthState {
	state := &healthState{}
	state.ready.Store(true)
	return state
}

func newHandler(cfg config.Config, logger *slog.Logger, health *healthState) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", proxy.NewRouter(cfg.Upstreams, cfg.Routes))
	mux.HandleFunc("GET /health/live", health.liveness)
	mux.HandleFunc("GET /health/ready", health.readiness)

	handler := http.Handler(mux)
	handler = middleware.AccessLog(logger, handler)
	return middleware.RequestID(handler)
}

func (h *healthState) liveness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *healthState) readiness(w http.ResponseWriter, r *http.Request) {
	if !h.ready.Load() {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
}
