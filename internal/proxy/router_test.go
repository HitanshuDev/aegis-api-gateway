package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/HitanshuDev/aegis-api-gateway/internal/config"
)

func TestRouterRoutesByLongestPathPrefix(t *testing.T) {
	usersPaths := make(chan string, 1)
	v2Paths := make(chan string, 1)
	usersServer := newUpstreamServer(t, "users", usersPaths)
	defer usersServer.Close()
	v2Server := newUpstreamServer(t, "v2", v2Paths)
	defer v2Server.Close()

	usersURL := parseURL(t, usersServer.URL)
	v2URL := parseURL(t, v2Server.URL)
	handler := NewRouter(
		map[string]*url.URL{
			"users": usersURL,
			"v2":    v2URL,
		},
		[]config.Route{
			{Prefix: "/api", Upstream: "users"},
			{Prefix: "/api/v2", Upstream: "v2"},
		},
	)

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
		wantPath   string
		wantTarget chan string
	}{
		{
			name:       "first route preserves path",
			path:       "/api/users/42",
			wantStatus: http.StatusOK,
			wantBody:   "users",
			wantPath:   "/api/users/42",
			wantTarget: usersPaths,
		},
		{
			name:       "longest overlapping prefix wins",
			path:       "/api/v2/items",
			wantStatus: http.StatusOK,
			wantBody:   "v2",
			wantPath:   "/api/v2/items",
			wantTarget: v2Paths,
		},
		{
			name:       "prefix boundary does not match",
			path:       "/apix/items",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "unmatched path returns not found",
			path:       "/other",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			handler.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if tt.wantStatus != http.StatusOK {
				return
			}
			if response.Body.String() != tt.wantBody {
				t.Fatalf("body = %q, want %q", response.Body.String(), tt.wantBody)
			}
			if got := <-tt.wantTarget; got != tt.wantPath {
				t.Errorf("upstream path = %q, want %q", got, tt.wantPath)
			}
		})
	}
}

func newUpstreamServer(t *testing.T, name string, paths chan<- string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths <- r.URL.Path
		_, _ = io.WriteString(w, name)
	}))
}

func parseURL(t *testing.T, rawURL string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
