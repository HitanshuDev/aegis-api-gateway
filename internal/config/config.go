package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
)

type Config struct {
	ListenAddr  string
	UpstreamURL *url.URL
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
		ListenAddr:  listenAddr,
		UpstreamURL: target,
	}, nil
}
