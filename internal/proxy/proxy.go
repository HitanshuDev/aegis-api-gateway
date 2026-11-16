package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"strings"

	"github.com/HitanshuDev/aegis-api-gateway/internal/config"
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

func NewRouter(upstreams map[string]*url.URL, routes []config.Route) http.Handler {
	type routedProxy struct {
		prefix  string
		handler http.Handler
	}

	orderedRoutes := append([]config.Route(nil), routes...)
	sort.Slice(orderedRoutes, func(i, j int) bool {
		return len(orderedRoutes[i].Prefix) > len(orderedRoutes[j].Prefix)
	})

	proxies := make([]routedProxy, 0, len(orderedRoutes))
	for _, route := range orderedRoutes {
		proxies = append(proxies, routedProxy{
			prefix:  route.Prefix,
			handler: New(upstreams[route.Upstream]),
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, route := range proxies {
			if route.prefix == "/" || r.URL.Path == route.prefix || strings.HasPrefix(r.URL.Path, route.prefix+"/") {
				route.handler.ServeHTTP(w, r)
				return
			}
		}

		http.NotFound(w, r)
	})
}
