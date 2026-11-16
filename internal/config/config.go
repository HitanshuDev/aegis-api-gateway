package config

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func Load(filename string) (Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return Config{}, fmt.Errorf("open config file %q: %w", filename, err)
	}
	defer file.Close()

	var yamlConfig YAMLConfig
	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)
	if err := decoder.Decode(&yamlConfig); err != nil {
		return Config{}, fmt.Errorf("decode config file %q: %w", filename, err)
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Config{}, fmt.Errorf("config file %q must contain one YAML document", filename)
		}
		return Config{}, fmt.Errorf("decode config file %q: %w", filename, err)
	}

	cfg, err := validate(yamlConfig)
	if err != nil {
		return Config{}, fmt.Errorf("invalid config file %q: %w", filename, err)
	}

	return cfg, nil
}

func validate(raw YAMLConfig) (Config, error) {
	if raw.ListenAddr == "" {
		return Config{}, fmt.Errorf("listen_addr is required")
	}
	if _, _, err := net.SplitHostPort(raw.ListenAddr); err != nil {
		return Config{}, fmt.Errorf("invalid listen_addr %q: %w", raw.ListenAddr, err)
	}

	readTimeout, err := parseTimeout("timeouts.read", raw.Timeouts.Read)
	if err != nil {
		return Config{}, err
	}
	readHeaderTimeout, err := parseTimeout("timeouts.read_header", raw.Timeouts.ReadHeader)
	if err != nil {
		return Config{}, err
	}
	writeTimeout, err := parseTimeout("timeouts.write", raw.Timeouts.Write)
	if err != nil {
		return Config{}, err
	}
	idleTimeout, err := parseTimeout("timeouts.idle", raw.Timeouts.Idle)
	if err != nil {
		return Config{}, err
	}
	shutdownTimeout, err := parseTimeout("timeouts.shutdown", raw.Timeouts.Shutdown)
	if err != nil {
		return Config{}, err
	}

	if len(raw.Upstreams) == 0 {
		return Config{}, fmt.Errorf("upstreams must define at least one upstream")
	}

	upstreams := make(map[string]*url.URL, len(raw.Upstreams))
	for name, upstream := range raw.Upstreams {
		if strings.TrimSpace(name) == "" {
			return Config{}, fmt.Errorf("upstreams contains an empty name")
		}
		target, err := url.Parse(upstream)
		if err != nil {
			return Config{}, fmt.Errorf("upstreams.%s is invalid: %w", name, err)
		}
		if target.Host == "" ||
			(target.Scheme != "http" && target.Scheme != "https") {
			return Config{}, fmt.Errorf(
				"upstreams.%s must be an absolute HTTP or HTTPS URL with a host",
				name,
			)
		}
		upstreams[name] = target
	}

	if len(raw.Routes) == 0 {
		return Config{}, fmt.Errorf("routes must define at least one path prefix")
	}

	seenPrefixes := make(map[string]struct{}, len(raw.Routes))
	for i, route := range raw.Routes {
		if route.Prefix == "" || !strings.HasPrefix(route.Prefix, "/") || path.Clean(route.Prefix) != route.Prefix || strings.ContainsAny(route.Prefix, "?#") {
			return Config{}, fmt.Errorf("routes[%d].prefix %q must be a canonical absolute path prefix", i, route.Prefix)
		}
		if _, exists := seenPrefixes[route.Prefix]; exists {
			return Config{}, fmt.Errorf("duplicate route prefix %q", route.Prefix)
		}
		seenPrefixes[route.Prefix] = struct{}{}

		if _, exists := upstreams[route.Upstream]; !exists {
			return Config{}, fmt.Errorf("routes[%d] prefix %q references undefined upstream %q", i, route.Prefix, route.Upstream)
		}
	}

	return Config{
		ListenAddr:        raw.ListenAddr,
		Upstreams:         upstreams,
		Routes:            raw.Routes,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ShutdownTimeout:   shutdownTimeout,
	}, nil
}

func parseTimeout(name, value string) (time.Duration, error) {
	if value == "" {
		return 0, fmt.Errorf("%s is required", name)
	}
	timeout, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", name, value, err)
	}
	if timeout <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", name)
	}
	return timeout, nil
}
