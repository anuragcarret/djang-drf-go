package settings

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadWithTags(t *testing.T) {
	// Setup env
	os.Setenv("DEBUG", "true")
	os.Setenv("SECRET_KEY", "super-secret-key-that-is-long-enough-32-chars")
	defer os.Unsetenv("DEBUG")
	defer os.Unsetenv("SECRET_KEY")

	s, err := Load("")
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	if s.Debug != true {
		t.Errorf("Expected Debug to be true from env, got %v", s.Debug)
	}
	if s.SecretKey != "super-secret-key-that-is-long-enough-32-chars" {
		t.Errorf("Expected SecretKey from env")
	}
}

func TestDefaultValues(t *testing.T) {
	os.Setenv("SECRET_KEY", "super-secret-key-that-is-long-enough-32-chars")
	defer os.Unsetenv("SECRET_KEY")

	s, err := Load("")
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	if s.REST.Pagination.PageSize != 20 {
		t.Errorf("Expected default PageSize 20, got %d", s.REST.Pagination.PageSize)
	}
}

func TestRequiredValidation(t *testing.T) {
	// SecretKey is required
	os.Unsetenv("SECRET_KEY")

	_, err := Load("")
	if err == nil {
		t.Error("Expected error for missing required SecretKey")
	}
}

func TestYAMLLoading(t *testing.T) {
	yamlContent := `
debug: true
secret_key: "yaml-secret-key-long-enough-to-pass-validation-32"
rest_framework:
  pagination:
    page_size: 50
`
	tmpfile, _ := os.CreateTemp("", "settings*.yaml")
	defer os.Remove(tmpfile.Name())
	tmpfile.Write([]byte(yamlContent))
	tmpfile.Close()

	s, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	if s.REST.Pagination.PageSize != 50 {
		t.Errorf("Expected PageSize 50 from YAML, got %d", s.REST.Pagination.PageSize)
	}
}

func TestCustomSettings(t *testing.T) {
	type MyConfig struct {
		APIKey     string `json:"api_key" env:"MY_API_KEY" default:"key-123"`
		MaxRetries int    `json:"max_retries" default:"3"`
	}

	myConfig := &MyConfig{}
	RegisterCustom("myapp", myConfig)

	os.Setenv("SECRET_KEY", "super-secret-key-that-is-long-enough-32-chars")
	defer os.Unsetenv("SECRET_KEY")

	jsonContent := `{
		"secret_key": "super-secret-key-that-is-long-enough-32-chars",
		"custom": {
			"myapp": {
				"max_retries": 10
			}
		}
	}`
	tmpfile, _ := os.CreateTemp("", "settings*.json")
	defer os.Remove(tmpfile.Name())
	tmpfile.Write([]byte(jsonContent))
	tmpfile.Close()

	_, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	cfg := GetCustom[*MyConfig]("myapp")
	if cfg.MaxRetries != 10 {
		t.Errorf("Expected MaxRetries 10 from custom JSON, got %d", cfg.MaxRetries)
	}
	if cfg.APIKey != "key-123" {
		t.Errorf("Expected APIKey default value, got %s", cfg.APIKey)
	}
}

func TestLoadChain(t *testing.T) {
	baseDir, _ := os.MkdirTemp("", "config")
	defer os.RemoveAll(baseDir)

	basePath := filepath.Join(baseDir, "settings")

	os.WriteFile(basePath+".json", []byte(`{"debug": false, "secret_key": "base-secret-key-32-chars-long-enough"}`), 0644)
	os.WriteFile(basePath+".production.json", []byte(`{"debug": true}`), 0644)

	s, err := LoadChain(basePath, "production")
	if err != nil {
		t.Fatalf("Failed to load chain: %v", err)
	}

	if s.Debug != true {
		t.Error("Production override failed")
	}
	if s.SecretKey != "base-secret-key-32-chars-long-enough" {
		t.Error("Base setting lost")
	}
}

func TestPoolConfigDuration(t *testing.T) {
	os.Setenv("SECRET_KEY", "super-secret-key-that-is-long-enough-32-chars")
	defer os.Unsetenv("SECRET_KEY")

	s, err := Load("")
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	if s.Databases == nil {
		s.Databases = make(map[string]DatabaseConfig)
	}
	// Note: currently processTags doesn't recurse into maps perfectly for mutation if not pointer
	// But let's check the base PoolConfig logic for defaults

	pc := &PoolConfig{}
	processTags(pc)

	if pc.MaxLifetime != 5*time.Minute {
		t.Errorf("Expected default MaxLifetime 5m, got %v", pc.MaxLifetime)
	}
}
