package hastrt

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMigrateLegacyConfigCopiesAicliTree(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	legacyConfig := filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "aicli", "hast")
	modernConfig := filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "sicli", "hast")
	if err := os.MkdirAll(legacyConfig, 0o700); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(legacyConfig, "config.toml")
	if err := os.WriteFile(configPath, []byte("base_url = \"https://backend.hast.so\"\n[auth]\nmode = \"key\"\napi_key = \"test-key\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := MigrateLegacyConfig(); err != nil {
		t.Fatalf("MigrateLegacyConfig() error = %v", err)
	}

	got, err := LoadFileConfig(filepath.Join(modernConfig, "config.toml"))
	if err != nil {
		t.Fatalf("LoadFileConfig() error = %v", err)
	}
	if got.Auth == nil || got.Auth.APIKey != "test-key" {
		t.Fatalf("migrated config = %#v", got)
	}
}

func TestMigrateLegacyConfigSkipsWhenDestinationConfigured(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	legacyConfig := filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "aicli", "hast")
	modernConfig := filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "sicli", "hast")
	for _, dir := range []string{legacyConfig, modernConfig} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(legacyConfig, "config.toml"), []byte("base_url = \"https://legacy.example.test\"\n[auth]\nmode = \"key\"\napi_key = \"legacy-key\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(modernConfig, "config.toml"), []byte("base_url = \"https://backend.hast.so\"\n[auth]\nmode = \"key\"\napi_key = \"modern-key\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := MigrateLegacyConfig(); err != nil {
		t.Fatalf("MigrateLegacyConfig() error = %v", err)
	}

	got, err := LoadFileConfig(filepath.Join(modernConfig, "config.toml"))
	if err != nil {
		t.Fatalf("LoadFileConfig() error = %v", err)
	}
	if got.Auth == nil || got.Auth.APIKey != "modern-key" {
		t.Fatalf("destination config overwritten: %#v", got)
	}
}
