package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/HitanshuDev/aegis-api-gateway/internal/config"
	"github.com/HitanshuDev/aegis-api-gateway/internal/gateway"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	handler := newHandler(cfg, logger)

	server := gateway.New(cfg.ListenAddr, handler)

	log.Printf("Aegis listening on %s, upstream=%s", cfg.ListenAddr, cfg.UpstreamURL)

	if err := server.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
