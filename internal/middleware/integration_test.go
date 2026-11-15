package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/HitanshuDev/aegis-api-gateway/internal/proxy"
)

func TestRequestIDPropagatedToUpstream(t *testing.T) {
	var upstreamRequestID string

	upstream := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			upstreamRequestID = r.Header.Get(RequestIDHeader)
			w.WriteHeader(http.StatusOK)
		}),
	)
	defer upstream.Close()

	target, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}

	proxyHandler := proxy.New(target)

	handler := RequestID(proxyHandler)

	gateway := httptest.NewServer(handler)
	defer gateway.Close()

	req, err := http.NewRequest(
		http.MethodGet,
		gateway.URL+"/users",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set(RequestIDHeader, "trace-123")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if got := resp.Header.Get(RequestIDHeader); got != "trace-123" {
		t.Fatalf("response request ID = %q", got)
	}

	if upstreamRequestID != "trace-123" {
		t.Fatalf("upstream request ID = %q", upstreamRequestID)
	}
}
