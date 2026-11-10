package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func New(target *url.URL) http.Handler {
	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(target)

			// Explicitly set the outgoing Host header to match the upstream target host
			pr.Out.Host = target.Host

			pr.Out.Header.Set("X-Request-ID", pr.In.Header.Get("X-Request-ID"))

			pr.SetXForwarded()
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("upstream proxy error: method=%s path=%s error=%v", r.Method, r.URL.Path, err)

			http.Error(w, "Bad Gateway", http.StatusBadGateway)
		},
	}

	return proxy
}
