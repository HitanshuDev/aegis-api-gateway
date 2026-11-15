package config

import (
	"testing"
	"time"
)

func TestLoadTimeoutDefaults(t *testing.T) {
	setBaseEnvironment(t)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	got := []time.Duration{
		cfg.ReadTimeout,
		cfg.ReadHeaderTimeout,
		cfg.WriteTimeout,
		cfg.IdleTimeout,
		cfg.ShutdownTimeout,
	}
	want := []time.Duration{
		15 * time.Second,
		5 * time.Second,
		30 * time.Second,
		60 * time.Second,
		30 * time.Second,
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("timeout %d = %s, want %s", i, got[i], want[i])
		}
	}
}

func TestLoadTimeoutOverrides(t *testing.T) {
	setBaseEnvironment(t)
	t.Setenv("AEGIS_READ_TIMEOUT", "2s")
	t.Setenv("AEGIS_READ_HEADER_TIMEOUT", "3s")
	t.Setenv("AEGIS_WRITE_TIMEOUT", "4s")
	t.Setenv("AEGIS_IDLE_TIMEOUT", "5s")
	t.Setenv("AEGIS_SHUTDOWN_TIMEOUT", "6s")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.ReadTimeout != 2*time.Second ||
		cfg.ReadHeaderTimeout != 3*time.Second ||
		cfg.WriteTimeout != 4*time.Second ||
		cfg.IdleTimeout != 5*time.Second ||
		cfg.ShutdownTimeout != 6*time.Second {
		t.Fatalf("unexpected timeout configuration: %+v", cfg)
	}
}

func TestLoadRejectsInvalidTimeout(t *testing.T) {
	setBaseEnvironment(t)
	t.Setenv("AEGIS_WRITE_TIMEOUT", "not-a-duration")

	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted an invalid timeout")
	}
}

func TestLoadRejectsNonPositiveTimeout(t *testing.T) {
	setBaseEnvironment(t)
	t.Setenv("AEGIS_SHUTDOWN_TIMEOUT", "0s")

	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted a zero shutdown timeout")
	}
}

func setBaseEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("AEGIS_LISTEN_ADDR", ":8080")
	t.Setenv("AEGIS_UPSTREAM_URL", "http://localhost:9000")
	t.Setenv("AEGIS_READ_TIMEOUT", "")
	t.Setenv("AEGIS_READ_HEADER_TIMEOUT", "")
	t.Setenv("AEGIS_WRITE_TIMEOUT", "")
	t.Setenv("AEGIS_IDLE_TIMEOUT", "")
	t.Setenv("AEGIS_SHUTDOWN_TIMEOUT", "")
}
