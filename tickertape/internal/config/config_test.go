package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadUsesEnvironmentCredentialWithoutPersistingIt(t *testing.T) {
	t.Setenv("TICKERTAPE_AUTH_HEADER", "Bearer test-token")
	t.Setenv("TICKERTAPE_COOKIE", "")
	t.Setenv("TICKERTAPE_USER_AGENT", "")
	path := filepath.Join(t.TempDir(), "config.json")
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AuthSource != "environment" || cfg.AuthHeader() != "Bearer test-token" || !cfg.CredentialConfigured() {
		t.Fatalf("unexpected auth resolution: source=%q configured=%v", cfg.AuthSource, cfg.CredentialConfigured())
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("Load unexpectedly persisted config at %s", path)
	}
}
