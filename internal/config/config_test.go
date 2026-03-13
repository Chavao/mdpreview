package config_test

import (
	"testing"

	"github.com/Chavao/mdpreview/internal/config"
)

func TestParseUsesDefaults(t *testing.T) {
	t.Parallel()

	cfg, err := config.Parse(nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if cfg.Host != config.DefaultHost {
		t.Fatalf("Host = %q, want %q", cfg.Host, config.DefaultHost)
	}

	if cfg.Port != config.DefaultPort {
		t.Fatalf("Port = %d, want %d", cfg.Port, config.DefaultPort)
	}
}

func TestParseOverridesDefaults(t *testing.T) {
	t.Parallel()

	cfg, err := config.Parse([]string{"--host", "127.0.0.1", "--port", "8080"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if cfg.Host != "127.0.0.1" {
		t.Fatalf("Host = %q, want %q", cfg.Host, "127.0.0.1")
	}

	if cfg.Port != 8080 {
		t.Fatalf("Port = %d, want %d", cfg.Port, 8080)
	}
}

func TestParseRejectsInvalidPort(t *testing.T) {
	t.Parallel()

	if _, err := config.Parse([]string{"--port", "0"}); err == nil {
		t.Fatal("Parse() error = nil, want error")
	}
}
