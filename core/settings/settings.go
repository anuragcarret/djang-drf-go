package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// Conf is the global configuration instance
var Conf *Settings

// Settings holds all framework and application configuration
type Settings struct {
	Debug         bool                      `json:"debug"`
	SecretKey     string                    `json:"secret_key"`
	AllowedHosts  []string                  `json:"allowed_hosts"`
	Databases     map[string]DatabaseConfig `json:"databases"`
	InstalledApps []string                  `json:"installed_apps"`
	Middleware    []string                  `json:"middleware"`
	REST          RESTConfig                `json:"rest_framework"`
	Custom        map[string]interface{}    `json:"custom"`
}

// DatabaseConfig mirrors Django's DATABASES setting
type DatabaseConfig struct {
	Engine   string            `json:"engine"`
	Name     string            `json:"name"`
	User     string            `json:"user"`
	Password string            `json:"password"`
	Host     string            `json:"host"`
	Port     int               `json:"port"`
	Pool     PoolConfig        `json:"pool"`
	Options  map[string]string `json:"options"`
}

// PoolConfig for connection pooling
type PoolConfig struct {
	MaxOpen     int           `json:"max_open"`
	MaxIdle     int           `json:"max_idle"`
	MaxLifetime time.Duration `json:"max_lifetime"`
}

// RESTConfig mirrors DRF settings
type RESTConfig struct {
	DefaultAuthentication []string         `json:"default_authentication"`
	DefaultPermission     []string         `json:"default_permission"`
	Pagination            PaginationConfig `json:"pagination"`
}

// PaginationConfig defaults
type PaginationConfig struct {
	PageSize           int    `json:"page_size"`
	PageSizeQueryParam string `json:"page_size_query_param"`
	MaxPageSize        int    `json:"max_page_size"`
}

// Initialize sets the global configuration
func Initialize(s *Settings) {
	Conf = s
}

// Get returns the global settings (panics if not initialized)
func Get() *Settings {
	if Conf == nil {
		panic("settings not initialized")
	}
	return Conf
}

// Load reads configuration from file and environment
func Load(configPath string) (*Settings, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Expand environment variables
	content := expandEnv(string(data))

	var s Settings
	if err := json.Unmarshal([]byte(content), &s); err != nil {
		return nil, fmt.Errorf("failed to parse settings: %w", err)
	}

	// Validate after loading
	if err := s.Validate(); err != nil {
		return nil, err
	}

	return &s, nil
}

// MustLoad panics on error
func MustLoad(configPath string) *Settings {
	s, err := Load(configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load settings: %v", err))
	}
	return s
}

// Validate checks all required settings and constraints
func (s *Settings) Validate() error {
	var errs []string

	if s.SecretKey == "" {
		errs = append(errs, "secret_key is required")
	} else if len(s.SecretKey) < 32 {
		errs = append(errs, "secret_key must be at least 32 characters")
	}

	if !s.Debug && len(s.AllowedHosts) == 0 {
		errs = append(errs, "allowed_hosts required when debug=false")
	}

	if len(errs) > 0 {
		return fmt.Errorf("settings validation failed: %s", strings.Join(errs, "; "))
	}

	return nil
}

// expandEnv replaces ${VAR} or ${VAR:default} in string
func expandEnv(s string) string {
	re := regexp.MustCompile(`\$\{([^}:]+)(?::([^}]+))?\}`)
	return re.ReplaceAllStringFunc(s, func(m string) string {
		match := re.FindStringSubmatch(m)
		key := match[1]
		def := ""
		if len(match) > 2 {
			def = match[2]
		}

		val := os.Getenv(key)
		if val == "" {
			return def
		}
		return val
	})
}
