package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HitanshuDev/aegis-api-gateway/internal/config"
	"github.com/HitanshuDev/aegis-api-gateway/internal/gateway"
)

func run(
	server *gateway.Server,
	health *healthState,
	shutdownTimeout time.Duration,
	signals <-chan os.Signal,
) error {
	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- server.Start()
	}()

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil

	case <-signals:
		health.ready.Store(false)

		ctx, cancel := context.WithTimeout(
			context.Background(),
			shutdownTimeout,
		)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			_ = server.Close()
			return err
		}

		err := <-serverErrors
		if err != nil && err != http.ErrServerClosed {
			return err
		}

		return nil
	}
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	health := newHealthState()
	handler := newHandler(cfg, logger, health)

	server := gateway.New(cfg.ListenAddr, handler, gateway.Timeouts{
		Read:       cfg.ReadTimeout,
		ReadHeader: cfg.ReadHeaderTimeout,
		Write:      cfg.WriteTimeout,
		Idle:       cfg.IdleTimeout,
	})

	log.Printf("Aegis listening on %s, upstream=%s", cfg.ListenAddr, cfg.UpstreamURL)

	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)

	if err := run(server, health, cfg.ShutdownTimeout, signals); err != nil {
		logger.Error("gateway stopped with error", "error", err)
		os.Exit(1)
	}

}
