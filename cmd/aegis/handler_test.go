package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/HitanshuDev/aegis-api-gateway/internal/config"
)

func TestHealthEndpoints(t *testing.T) {
	upstream, err := url.Parse("http://localhost:9000")
	if err != nil {
		t.Fatal(err)
	}

	health := newHealthState()
	handler := newHandler(
		config.Config{
			Upstreams: map[string]*url.URL{"default": upstream},
			Routes:    []config.Route{{Prefix: "/", Upstream: "default"}},
		},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		health,
	)

	tests := []struct {
		name       string
		method     string
		path       string
		ready      bool
		wantStatus int
	}{
		{
			name:       "liveness",
			method:     http.MethodGet,
			path:       "/health/live",
			ready:      true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "liveness head",
			method:     http.MethodHead,
			path:       "/health/live",
			ready:      true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "readiness",
			method:     http.MethodGet,
			path:       "/health/ready",
			ready:      true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "not ready",
			method:     http.MethodGet,
			path:       "/health/ready",
			ready:      false,
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "unsupported method rejected by mux",
			method:     http.MethodPost,
			path:       "/health/live",
			ready:      true,
			wantStatus: http.StatusBadGateway,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			health.ready.Store(tt.ready)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if tt.wantStatus == http.StatusMethodNotAllowed &&
				recorder.Header().Get("Allow") != http.MethodGet {
				t.Errorf("Allow = %q, want %q",
					recorder.Header().Get("Allow"), http.MethodGet)
			}
		})
	}
}
