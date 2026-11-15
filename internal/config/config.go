package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"time"
)

type Config struct {
	ListenAddr        string
	UpstreamURL       *url.URL
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
}

func Load() (Config, error) {
	listenAddr := os.Getenv("AEGIS_LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":8080"
	}

	upstream := os.Getenv("AEGIS_UPSTREAM_URL")
	if upstream == "" {
		upstream = "http://localhost:9000"
	}

	readTimeout, err := durationFromEnv("AEGIS_READ_TIMEOUT", 15*time.Second)
	if err != nil {
		return Config{}, err
	}

	readHeaderTimeout, err := durationFromEnv(
		"AEGIS_READ_HEADER_TIMEOUT",
		5*time.Second,
	)
	if err != nil {
		return Config{}, err
	}

	writeTimeout, err := durationFromEnv("AEGIS_WRITE_TIMEOUT", 30*time.Second)
	if err != nil {
		return Config{}, err
	}

	idleTimeout, err := durationFromEnv("AEGIS_IDLE_TIMEOUT", 60*time.Second)
	if err != nil {
		return Config{}, err
	}

	shutdownTimeout, err := durationFromEnv(
		"AEGIS_SHUTDOWN_TIMEOUT",
		30*time.Second,
	)
	if err != nil {
		return Config{}, err
	}

	if _, _, err := net.SplitHostPort(listenAddr); err != nil {
		return Config{}, fmt.Errorf("invalid listen address: %w", err)
	}

	target, err := url.Parse(upstream)
	if err != nil {
		return Config{}, fmt.Errorf("parse upstream URL: %w", err)
	}

	if target.Host == "" ||
		(target.Scheme != "http" && target.Scheme != "https") {
		return Config{}, fmt.Errorf("upstream must be an absolute HTTP or HTTPS URL")
	}

	return Config{
		ListenAddr:        listenAddr,
		UpstreamURL:       target,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ShutdownTimeout:   shutdownTimeout,
	}, nil
}

func durationFromEnv(name string, defaultValue time.Duration) (time.Duration, error) {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", name, err)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", name)
	}

	return duration, nil
}
