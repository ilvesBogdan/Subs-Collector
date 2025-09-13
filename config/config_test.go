package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const (
	// Keys
	keyDatabaseURL = "DATABASE_URL"
	keyPort        = "PORT"

	// values .env
	envDBURL = "postgres://env_user:env_pass@localhost:5432/envdb"
	envPort  = "1234"

	// values YAML
	yamlDBURL = "postgres://yaml_user:yaml_pass@localhost:5432/yamldb"
	yamlPort  = "5678"

	// values env
	osEnvDBURL = "postgres://envvar_user:envvar_pass@localhost:5432/envvardb"
	osEnvPort  = "9999"

	// values default
	defaultPort = "8080"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
	return path
}

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// .env
	dotEnvContent := fmt.Sprintf("%s=%s\n%s=%s\n", keyDatabaseURL, envDBURL, keyPort, envPort)
	dotEnvPath := writeFile(t, tmpDir, ".env", dotEnvContent)

	yamlContent := fmt.Sprintf("%s: %s\n%s: %s\n", keyDatabaseURL, yamlDBURL, keyPort, yamlPort)
	yamlPath := writeFile(t, tmpDir, "config.yaml", yamlContent)

	cfg := Load(dotEnvPath, yamlPath)
	if cfg.DatabaseURL != envDBURL {
		t.Errorf("expected %s from .env, got %s", keyDatabaseURL, cfg.DatabaseURL)
	}
	if cfg.Port != envPort {
		t.Errorf("expected %s=%s from .env, got %s", keyPort, envPort, cfg.Port)
	}

	// YAML
	emptyDotEnv := writeFile(t, tmpDir, ".env", "")
	cfg = Load(emptyDotEnv, yamlPath)
	if cfg.DatabaseURL != yamlDBURL {
		t.Errorf("expected %s from yaml, got %s", keyDatabaseURL, cfg.DatabaseURL)
	}
	if cfg.Port != yamlPort {
		t.Errorf("expected %s=%s from yaml, got %s", keyPort, yamlPort, cfg.Port)
	}

	// env
	_ = os.Setenv(keyDatabaseURL, osEnvDBURL)
	_ = os.Setenv(keyPort, osEnvPort)
	defer func() { _ = os.Unsetenv(keyDatabaseURL) }()
	defer func() { _ = os.Unsetenv(keyPort) }()

	cfg = Load(emptyDotEnv, filepath.Join(tmpDir, "nonexistent.yaml"))
	if cfg.DatabaseURL != osEnvDBURL {
		t.Errorf("expected %s from env, got %s", keyDatabaseURL, cfg.DatabaseURL)
	}
	if cfg.Port != osEnvPort {
		t.Errorf("expected %s=%s from env, got %s", keyPort, osEnvPort, cfg.Port)
	}

	// default
	_ = os.Unsetenv(keyPort)
	cfg = Load(emptyDotEnv, filepath.Join(tmpDir, "nonexistent.yaml"))
	if cfg.Port != defaultPort {
		t.Errorf("expected default %s=%s, got %s", keyPort, defaultPort, cfg.Port)
	}

	// snake_case
	yamlSnake := fmt.Sprintf("database_url: %s\n", yamlDBURL)
	yamlSnakePath := writeFile(t, tmpDir, "config_snake.yaml", yamlSnake)
	cfg = Load(emptyDotEnv, yamlSnakePath)
	if cfg.DatabaseURL != yamlDBURL {
		t.Errorf("expected %s from snake_case, got %s", yamlDBURL, cfg.DatabaseURL)
	}

	// kebab-case
	yamlKebab := fmt.Sprintf("database-url: %s\n", yamlDBURL)
	yamlKebabPath := writeFile(t, tmpDir, "config_kebab.yaml", yamlKebab)
	cfg = Load(emptyDotEnv, yamlKebabPath)
	if cfg.DatabaseURL != yamlDBURL {
		t.Errorf("expected %s from kebab-case, got %s", yamlDBURL, cfg.DatabaseURL)
	}

	// panic
	_ = os.Unsetenv(keyDatabaseURL)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for missing %s, but no panic occurred", keyDatabaseURL)
		}
	}()
	Load(filepath.Join(tmpDir, "no.env"), filepath.Join(tmpDir, "no.yaml"))
}
