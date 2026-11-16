package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const validYAML = `listen_addr: ":8080"
timeouts:
  read: 15s
  read_header: 5s
  write: 30s
  idle: 60s
  shutdown: 10s
upstreams:
  users: http://localhost:9001
  orders: http://localhost:9002
routes:
  - prefix: /users
    upstream: users
  - prefix: /orders
    upstream: orders
`

func TestLoadYAMLConfiguration(t *testing.T) {
	cfg := loadConfigText(t, validYAML)

	if cfg.ListenAddr != ":8080" {
		t.Errorf("ListenAddr = %q, want %q", cfg.ListenAddr, ":8080")
	}
	if cfg.ReadTimeout != 15*time.Second ||
		cfg.ReadHeaderTimeout != 5*time.Second ||
		cfg.WriteTimeout != 30*time.Second ||
		cfg.IdleTimeout != 60*time.Second ||
		cfg.ShutdownTimeout != 10*time.Second {
		t.Errorf("unexpected timeout values: %+v", cfg)
	}
	if len(cfg.Upstreams) != 2 || len(cfg.Routes) != 2 {
		t.Fatalf("got %d upstreams and %d routes, want 2 of each",
			len(cfg.Upstreams), len(cfg.Routes))
	}
	if got := cfg.Upstreams["users"].String(); got != "http://localhost:9001" {
		t.Errorf("users upstream = %q", got)
	}
}

func TestExampleYAMLConfigurationLoads(t *testing.T) {
	if _, err := Load(filepath.Join("..", "..", "config.yaml")); err != nil {
		t.Fatal(err)
	}
}

func TestLoadRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		wantMessage string
	}{
		{
			name:        "invalid upstream URL",
			yaml:        strings.Replace(validYAML, "http://localhost:9001", "ftp://localhost:9001", 1),
			wantMessage: "upstreams.users must be an absolute HTTP or HTTPS URL",
		},
		{
			name:        "duplicate prefix",
			yaml:        strings.Replace(validYAML, "  - prefix: /orders", "  - prefix: /users", 1),
			wantMessage: `duplicate route prefix "/users"`,
		},
		{
			name:        "unknown upstream",
			yaml:        strings.Replace(validYAML, "upstream: orders", "upstream: missing", 1),
			wantMessage: `references undefined upstream "missing"`,
		},
		{
			name:        "noncanonical prefix",
			yaml:        strings.Replace(validYAML, "prefix: /users", "prefix: /users/", 1),
			wantMessage: "must be a canonical absolute path prefix",
		},
		{
			name:        "invalid timeout",
			yaml:        strings.Replace(validYAML, "read: 15s", "read: sometime", 1),
			wantMessage: "invalid timeouts.read",
		},
		{
			name:        "unknown field",
			yaml:        strings.Replace(validYAML, "listen_addr:", "listen_address:", 1),
			wantMessage: "field listen_address not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loadConfigTextWithError(t, tt.yaml)
			if err == nil {
				t.Fatal("Load() accepted invalid configuration")
			}
			if !strings.Contains(err.Error(), tt.wantMessage) {
				t.Fatalf("Load() error = %q, want it to contain %q", err, tt.wantMessage)
			}
		})
	}
}

func TestLoadMissingFileIncludesPath(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "missing.yaml")
	_, err := Load(filename)
	if err == nil {
		t.Fatal("Load() accepted a missing file")
	}
	if !strings.Contains(err.Error(), filename) {
		t.Errorf("error %q does not identify filename %q", err, filename)
	}
}

func loadConfigText(t *testing.T, text string) Config {
	t.Helper()
	cfg, err := loadConfigTextWithError(t, text)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func loadConfigTextWithError(t *testing.T, text string) (Config, error) {
	t.Helper()
	filename := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(filename, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	return Load(filename)
}
