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
			// Set the target URL (automatically handles path, query, and upstream URL structure)
			pr.SetURL(target)

			// Explicitly set the outgoing Host header to match the upstream target host
			pr.Out.Host = target.Host

			// Optional: If you want to preserve standard X-Forwarded-* headers (like X-Forwarded-For)
			pr.SetXForwarded()
		},
		ErrorHandler: func(
			w http.ResponseWriter,
			r *http.Request,
			err error,
		) {
			log.Printf(
				"upstream proxy error: method=%s path=%s error=%v",
				r.Method,
				r.URL.Path,
				err,
			)

			http.Error(
				w,
				"Bad Gateway",
				http.StatusBadGateway,
			)
		},
	}

	return proxy
}
