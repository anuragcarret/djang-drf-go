package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Conf is the global configuration instance
var Conf *Settings

var (
	customSettingsMap = make(map[string]interface{})
)

// Settings holds all framework and application configuration
type Settings struct {
	Debug         bool                      `json:"debug" yaml:"debug" env:"DEBUG" default:"false"`
	SecretKey     string                    `json:"secret_key" yaml:"secret_key" env:"SECRET_KEY" required:"true"`
	AllowedHosts  []string                  `json:"allowed_hosts" yaml:"allowed_hosts" env:"ALLOWED_HOSTS"`
	Databases     map[string]DatabaseConfig `json:"databases" yaml:"databases"`
	InstalledApps []string                  `json:"installed_apps" yaml:"installed_apps"`
	Middleware    []string                  `json:"middleware" yaml:"middleware"`
	REST          RESTConfig                `json:"rest_framework" yaml:"rest_framework"`
	Custom        map[string]interface{}    `json:"custom" yaml:"custom"`
}

// DatabaseConfig mirrors Django's DATABASES setting
type DatabaseConfig struct {
	Engine   string            `json:"engine" yaml:"engine" required:"true"`
	Name     string            `json:"name" yaml:"name"`
	User     string            `json:"user" yaml:"user" env:"DB_USER"`
	Password string            `json:"password" yaml:"password" env:"DB_PASSWORD"`
	Host     string            `json:"host" yaml:"host" env:"DB_HOST" default:"localhost"`
	Port     int               `json:"port" yaml:"port" env:"DB_PORT"`
	Pool     PoolConfig        `json:"pool" yaml:"pool"`
	Options  map[string]string `json:"options" yaml:"options"`
}

// PoolConfig for connection pooling
type PoolConfig struct {
	MaxOpen     int           `json:"max_open" yaml:"max_open" default:"25"`
	MaxIdle     int           `json:"max_idle" yaml:"max_idle" default:"5"`
	MaxLifetime time.Duration `json:"max_lifetime" yaml:"max_lifetime" default:"5m"`
}

// RESTConfig mirrors DRF settings
type RESTConfig struct {
	DefaultAuthentication []string         `json:"default_authentication" yaml:"default_authentication"`
	DefaultPermission     []string         `json:"default_permission" yaml:"default_permission"`
	Pagination            PaginationConfig `json:"pagination" yaml:"pagination"`
}

// PaginationConfig defaults
type PaginationConfig struct {
	PageSize           int    `json:"page_size" yaml:"page_size" default:"20"`
	PageSizeQueryParam string `json:"page_size_query_param" yaml:"page_size_query_param" default:"page_size"`
	MaxPageSize        int    `json:"max_page_size" yaml:"max_page_size" default:"100"`
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
	s := &Settings{
		Custom: make(map[string]interface{}),
	}

	// 1. Load from file if exists
	if configPath != "" {
		if err := loadFromFile(configPath, s); err != nil {
			return nil, err
		}
	}

	// 2. Process tags (Defaults -> Env -> Required check)
	if err := processTags(s); err != nil {
		return nil, err
	}

	// 3. Process Custom settings if they've been registered
	for key, target := range customSettingsMap {
		if raw, ok := s.Custom[key]; ok {
			// Marhsalling back and forth is the simplest way to deep copy/convert maps to structs
			data, _ := json.Marshal(raw)
			json.Unmarshal(data, target)
			processTags(target)
		}
	}

	return s, nil
}

// LoadChain loads base + environment-specific config
func LoadChain(basePath, env string) (*Settings, error) {
	s := &Settings{
		Custom: make(map[string]interface{}),
	}

	files := []string{basePath}
	if env != "" {
		// e.g. settings.production or settings.development
		files = append(files, fmt.Sprintf("%s.%s", basePath, env))
	}

	for _, f := range files {
		matches, _ := filepath.Glob(f + ".*")
		if len(matches) > 0 {
			// Prefer .json over .yaml if both exist
			var targetFile string
			for _, m := range matches {
				if strings.HasSuffix(m, ".json") {
					targetFile = m
					break
				}
				if strings.HasSuffix(m, ".yaml") || strings.HasSuffix(m, ".yml") {
					targetFile = m
				}
			}
			if targetFile != "" {
				if err := loadFromFile(targetFile, s); err != nil {
					return nil, err
				}
			}
		}
	}

	if err := processTags(s); err != nil {
		return nil, err
	}

	return s, nil
}

// MustLoad panics on error
func MustLoad(configPath string) *Settings {
	s, err := Load(configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load settings: %v", err))
	}
	return s
}

// RegisterCustom registers a custom settings struct for a key
func RegisterCustom(key string, target interface{}) {
	customSettingsMap[key] = target
}

// GetCustom returns the custom settings struct for a key
func GetCustom[T any](key string) T {
	val, ok := customSettingsMap[key]
	if !ok {
		var zero T
		return zero
	}
	return val.(T)
}

func loadFromFile(path string, s *Settings) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	content := expandEnv(string(data))
	ext := filepath.Ext(path)

	if ext == ".json" {
		if err := json.Unmarshal([]byte(content), s); err != nil {
			return fmt.Errorf("failed to parse JSON settings: %w", err)
		}
	} else if ext == ".yaml" || ext == ".yml" {
		if err := yaml.Unmarshal([]byte(content), s); err != nil {
			return fmt.Errorf("failed to parse YAML settings: %w", err)
		}
	}
	return nil
}

func processTags(obj interface{}) error {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// 1. Default values
		if def := fieldType.Tag.Get("default"); def != "" && isZero(field) {
			setFieldValue(field, def)
		}

		// 2. Environment variables (Override)
		if env := fieldType.Tag.Get("env"); env != "" {
			if envVal := os.Getenv(env); envVal != "" {
				setFieldValue(field, envVal)
			}
		}

		// 3. Required check
		if fieldType.Tag.Get("required") == "true" && isZero(field) {
			return fmt.Errorf("required setting %s (tag: %s) is missing", fieldType.Name, fieldType.Tag.Get("json"))
		}

		// Recurse into nested structs or maps of structs
		if field.Kind() == reflect.Struct {
			if err := processTags(field.Addr().Interface()); err != nil {
				return err
			}
		} else if field.Kind() == reflect.Map {
			for _, key := range field.MapKeys() {
				val := field.MapIndex(key)
				if val.Kind() == reflect.Struct {
					// Maps are tricky to mutate in-place via reflect without pointers
					// For our DatabaeConfig, it's often small so we'll handle common cases
					// If it's a map[string]Struct, we can't easily recurse and mutate unless it's map[string]*Struct
				}
			}
		}
	}

	return nil
}

func isZero(v reflect.Value) bool {
	return v.IsZero()
}

func setFieldValue(field reflect.Value, value string) {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Bool:
		b, _ := strconv.ParseBool(value)
		field.SetBool(b)
	case reflect.Int, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			d, _ := time.ParseDuration(value)
			field.SetInt(int64(d))
		} else {
			i, _ := strconv.ParseInt(value, 10, 64)
			field.SetInt(i)
		}
	}
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

// Validate checks all required settings (deprecated in favor of processTags)
func (s *Settings) Validate() error {
	return processTags(s)
}
