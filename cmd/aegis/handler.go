package main

import (
	"log/slog"
	"net/http"

	"github.com/HitanshuDev/aegis-api-gateway/internal/config"
	"github.com/HitanshuDev/aegis-api-gateway/internal/middleware"
	"github.com/HitanshuDev/aegis-api-gateway/internal/proxy"
)

func newHandler(cfg config.Config, logger *slog.Logger) http.Handler {
	handler := proxy.New(cfg.UpstreamURL)
	handler = middleware.AccessLog(logger, cfg.UpstreamURL.Host, handler)
	return middleware.RequestID(handler)
}
