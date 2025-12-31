package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("loads settings from JSON file", func(t *testing.T) {
		content := `{
			"debug": true,
			"secret_key": "test-secret-key-at-least-32-chars-long",
			"databases": {
				"default": {
					"engine": "postgres",
					"name": "testdb"
				}
			}
		}`
		tmpFile := filepath.Join(t.TempDir(), "settings.json")
		err := os.WriteFile(tmpFile, []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}

		s, err := Load(tmpFile)
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if s.Debug != true {
			t.Error("expected Debug to be true")
		}
		if s.SecretKey != "test-secret-key-at-least-32-chars-long" {
			t.Error("SecretKey mismatch")
		}
		if s.Databases["default"].Engine != "postgres" {
			t.Error("Database engine mismatch")
		}
	})

	t.Run("loads with environment variable expansion", func(t *testing.T) {
		os.Setenv("TEST_SECRET", "env-secret-key-32-chars-long-minimal")
		defer os.Unsetenv("TEST_SECRET")

		content := `{
			"debug": true,
			"secret_key": "${TEST_SECRET}"
		}`
		tmpFile := filepath.Join(t.TempDir(), "settings.json")
		err := os.WriteFile(tmpFile, []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}

		s, err := Load(tmpFile)
		if err != nil {
			t.Fatal(err)
		}

		if s.SecretKey != "env-secret-key-32-chars-long-minimal" {
			t.Errorf("expected expanded env var, got %s", s.SecretKey)
		}
	})

	t.Run("fails on missing required setting", func(t *testing.T) {
		content := `{
			"debug": true
		}`
		tmpFile := filepath.Join(t.TempDir(), "settings.json")
		_ = os.WriteFile(tmpFile, []byte(content), 0644)

		_, err := Load(tmpFile)
		if err == nil {
			t.Error("expected error for missing required secret_key")
		}
	})
}

func TestValidate(t *testing.T) {
	t.Run("valid settings with debug", func(t *testing.T) {
		s := &Settings{
			Debug:     true,
			SecretKey: "a-very-long-and-secure-secret-key-32-chars",
			Databases: map[string]DatabaseConfig{
				"default": {Engine: "postgres", Name: "db"},
			},
		}
		err := s.Validate()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("valid settings without debug", func(t *testing.T) {
		s := &Settings{
			Debug:        false,
			SecretKey:    "a-very-long-and-secure-secret-key-32-chars",
			AllowedHosts: []string{"localhost"},
			Databases: map[string]DatabaseConfig{
				"default": {Engine: "postgres", Name: "db"},
			},
		}
		err := s.Validate()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("invalid without secret key", func(t *testing.T) {
		s := &Settings{
			Databases: map[string]DatabaseConfig{
				"default": {Engine: "postgres", Name: "db"},
			},
		}
		err := s.Validate()
		if err == nil {
			t.Error("expected validation error for missing secret key")
		}
	})
}

func TestGlobalAccess(t *testing.T) {
	s := &Settings{Debug: true}
	Initialize(s)

	if Get() != s {
		t.Error("Get() did not return initialized settings")
	}

	if Conf != s {
		t.Error("Conf global variable mismatch")
	}
}
