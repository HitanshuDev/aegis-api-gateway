package config

import (
	"net/url"
	"time"
)

type YAMLConfig struct {
	ListenAddr string            `yaml:"listen_addr"`
	Timeouts   YAMLTimeouts      `yaml:"timeouts"`
	Upstreams  map[string]string `yaml:"upstreams"`
	Routes     []Route           `yaml:"routes"`
}

type YAMLTimeouts struct {
	Read       string `yaml:"read"`
	ReadHeader string `yaml:"read_header"`
	Write      string `yaml:"write"`
	Idle       string `yaml:"idle"`
	Shutdown   string `yaml:"shutdown"`
}

type Route struct {
	Prefix   string `yaml:"prefix"`
	Upstream string `yaml:"upstream"`
}

type Config struct {
	ListenAddr        string
	Upstreams         map[string]*url.URL
	Routes            []Route
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
}
