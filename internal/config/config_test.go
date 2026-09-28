package config

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	cfg, err := Parse([]string{
		"--urls", "https://example.com, http://example.org",
		"--depth", "2",
		"--timeout", "30s",
		"--request-timeout", "2s",
		"--output", "result.json",
		"--log", "crawler.log",
	})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(cfg.URLs) != 2 || cfg.URLs[1] != "http://example.org" {
		t.Fatalf("unexpected URLs: %#v", cfg.URLs)
	}
	if cfg.MaxDepth != 2 {
		t.Fatalf("depth = %d, want 2", cfg.MaxDepth)
	}
}

func TestParseRejectsUnsupportedScheme(t *testing.T) {
	_, err := Parse([]string{"--urls", "ftp://example.com"})
	if err == nil || !strings.Contains(err.Error(), "http or https") {
		t.Fatalf("Parse() error = %v, want unsupported scheme error", err)
	}
}

func TestParseRejectsEmptyURLItem(t *testing.T) {
	_, err := Parse([]string{"--urls", "https://example.com,,https://example.org"})
	if err == nil {
		t.Fatal("Parse() error = nil, want empty URL item error")
	}
}
