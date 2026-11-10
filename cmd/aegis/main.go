package main

import (
	"log"
	"net/http"

	"github.com/HitanshuDev/aegis-api-gateway/internal/config"
	"github.com/HitanshuDev/aegis-api-gateway/internal/gateway"
	"github.com/HitanshuDev/aegis-api-gateway/internal/proxy"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	proxyHandler := proxy.New(cfg.UpstreamURL)

	server := gateway.New(cfg.ListenAddr, proxyHandler)

	log.Printf("Aegis listening on %s, upstream=%s", cfg.ListenAddr, cfg.UpstreamURL)

	if err := server.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
