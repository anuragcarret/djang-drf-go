# PRD: Core Settings System

> **Module:** `core/settings`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **Django Equivalent:** `django.conf.settings`

---

## 1. Overview

### 1.1 Purpose

The Settings System provides centralized configuration management for Django-DRF-Go applications:

- **Configuration Loading**: From files (JSON, YAML, TOML) and environment variables
- **Type-Safe Access**: Compile-time type checking for settings
- **Environment Support**: Development, staging, production configurations
- **Validation**: Ensure required settings are present and valid
- **Hot Reload**: Optional configuration refresh without restart (development)

### 1.2 Design Philosophy

| Django Approach | Go-Idiomatic Approach |
|-----------------|----------------------|
| Python module (`settings.py`) | Struct-based configuration |
| Dynamic attribute access | Type-safe struct fields |
| `getattr(settings, 'DEBUG', False)` | `settings.Debug` with Go defaults |
| `django.conf.settings` singleton | `Config` struct with global accessor |
| String-based lookups | Compile-time field access |

---

## 2. Core Concepts

### 2.1 Settings Structure

```go
// Settings holds all framework and application configuration
type Settings struct {
    // Debug enables debug mode (detailed errors, etc.)
    Debug bool `json:"debug" env:"DEBUG" default:"false"`
    
    // SecretKey for cryptographic signing
    SecretKey string `json:"secret_key" env:"SECRET_KEY" required:"true"`
    
    // AllowedHosts restricts which hosts can serve the app
    AllowedHosts []string `json:"allowed_hosts" env:"ALLOWED_HOSTS"`
    
    // Database configuration
    Databases map[string]DatabaseConfig `json:"databases"`
    
    // InstalledApps lists app names (for documentation; actual registration via imports)
    InstalledApps []string `json:"installed_apps"`
    
    // Middleware chain in execution order
    Middleware []string `json:"middleware"`
    
    // Templates configuration
    Templates TemplatesConfig `json:"templates"`
    
    // StaticFiles configuration
    StaticFiles StaticConfig `json:"static"`
    
    // REST Framework settings
    REST RESTConfig `json:"rest_framework"`
    
    // Logging configuration
    Logging LoggingConfig `json:"logging"`
    
    // Custom application settings
    Custom map[string]interface{} `json:"custom"`
}
```

### 2.2 Database Configuration

```go
// DatabaseConfig mirrors Django's DATABASES setting
type DatabaseConfig struct {
    // Engine: "postgres", "mysql", "sqlite3"
    Engine string `json:"engine" env:"DB_ENGINE" required:"true"`
    
    // Connection details
    Name     string `json:"name" env:"DB_NAME"`
    User     string `json:"user" env:"DB_USER"`
    Password string `json:"password" env:"DB_PASSWORD"`
    Host     string `json:"host" env:"DB_HOST" default:"localhost"`
    Port     int    `json:"port" env:"DB_PORT"`
    
    // Connection pool settings
    Pool PoolConfig `json:"pool"`
    
    // Options for additional driver-specific settings
    Options map[string]string `json:"options"`
}

// PoolConfig for connection pooling
type PoolConfig struct {
    MaxOpen     int           `json:"max_open" default:"25"`
    MaxIdle     int           `json:"max_idle" default:"5"`
    MaxLifetime time.Duration `json:"max_lifetime" default:"5m"`
}
```

### 2.3 REST Framework Configuration

```go
// RESTConfig mirrors DRF settings
type RESTConfig struct {
    // DefaultRendererClasses for content negotiation
    DefaultRendererClasses []string `json:"default_renderer_classes"`
    
    // DefaultParserClasses for request parsing
    DefaultParserClasses []string `json:"default_parser_classes"`
    
    // DefaultAuthentication classes
    DefaultAuthentication []string `json:"default_authentication"`
    
    // DefaultPermission classes
    DefaultPermission []string `json:"default_permission"`
    
    // Pagination settings
    Pagination PaginationConfig `json:"pagination"`
    
    // DefaultThrottleClasses for rate limiting
    DefaultThrottleClasses []string `json:"default_throttle_classes"`
    
    // ThrottleRates map throttle names to rates
    ThrottleRates map[string]string `json:"throttle_rates"`
    
    // ExceptionHandler function name
    ExceptionHandler string `json:"exception_handler"`
}

// PaginationConfig defaults
type PaginationConfig struct {
    PageSize          int    `json:"page_size" default:"20"`
    PageSizeQueryParam string `json:"page_size_query_param" default:"page_size"`
    MaxPageSize       int    `json:"max_page_size" default:"100"`
}
```

---

## 3. API Specification

### 3.1 Loading Settings

```go
// Load reads configuration from file and environment
func Load(configPath string) (*Settings, error)

// LoadWithEnv loads with explicit environment name
func LoadWithEnv(configPath, env string) (*Settings, error)

// MustLoad panics on error (use in main.go)
func MustLoad(configPath string) *Settings
```

### 3.2 Global Access

```go
// Package-level settings instance
var Conf *Settings

// Initialize sets the global configuration
func Initialize(s *Settings) {
    Conf = s
}

// Get returns the global settings (panics if not initialized)
func Get() *Settings
```

### 3.3 Settings Validation

```go
// Validate checks all required settings and constraints
func (s *Settings) Validate() error

// ValidateDatabase ensures database configuration is complete
func (s *Settings) ValidateDatabase(name string) error
```

### 3.4 Environment-Specific Loading

```go
// LoadChain loads base + environment-specific config
// e.g., settings.json + settings.production.json
func LoadChain(basePath, env string) (*Settings, error)

// Example:
// settings := settings.LoadChain("config/settings", "production")
// Loads: config/settings.json, then merges config/settings.production.json
```

---

## 4. Configuration File Formats

### 4.1 JSON Format (Default)

```json
{
  "debug": false,
  "secret_key": "${SECRET_KEY}",
  "allowed_hosts": ["example.com", "*.example.com"],
  
  "databases": {
    "default": {
      "engine": "postgres",
      "name": "${DB_NAME}",
      "user": "${DB_USER}",
      "password": "${DB_PASSWORD}",
      "host": "${DB_HOST:localhost}",
      "port": 5432,
      "pool": {
        "max_open": 25,
        "max_idle": 5
      }
    }
  },
  
  "middleware": [
    "security.SecurityMiddleware",
    "sessions.SessionMiddleware",
    "common.CommonMiddleware",
    "csrf.CSRFMiddleware",
    "auth.AuthenticationMiddleware"
  ],
  
  "rest_framework": {
    "default_authentication": ["TokenAuthentication"],
    "default_permission": ["IsAuthenticated"],
    "pagination": {
      "page_size": 20,
      "max_page_size": 100
    }
  }
}
```

### 4.2 YAML Format

```yaml
debug: false
secret_key: ${SECRET_KEY}
allowed_hosts:
  - example.com
  - "*.example.com"

databases:
  default:
    engine: postgres
    name: ${DB_NAME}
    user: ${DB_USER}
    password: ${DB_PASSWORD}
    host: ${DB_HOST:localhost}
    port: 5432

rest_framework:
  default_authentication:
    - TokenAuthentication
  pagination:
    page_size: 20
```

### 4.3 Environment Variable Syntax

| Syntax | Description | Example |
|--------|-------------|---------|
| `${VAR}` | Required variable | `${SECRET_KEY}` |
| `${VAR:default}` | With default value | `${DEBUG:false}` |
| `${VAR:-}` | Empty string default | `${OPTIONAL:-}` |

---

## 5. Django Comparison

### 5.1 Settings Definition

**Django (settings.py):**
```python
# settings.py
import os

DEBUG = os.environ.get('DEBUG', 'false').lower() == 'true'
SECRET_KEY = os.environ['SECRET_KEY']

DATABASES = {
    'default': {
        'ENGINE': 'django.db.backends.postgresql',
        'NAME': os.environ.get('DB_NAME', 'mydb'),
        'USER': os.environ.get('DB_USER', 'postgres'),
        'PASSWORD': os.environ['DB_PASSWORD'],
        'HOST': os.environ.get('DB_HOST', 'localhost'),
        'PORT': 5432,
    }
}

REST_FRAMEWORK = {
    'DEFAULT_AUTHENTICATION_CLASSES': [
        'rest_framework.authentication.TokenAuthentication',
    ],
    'PAGE_SIZE': 20,
}
```

**Django-DRF-Go (settings.json + main.go):**
```go
// main.go
package main

import (
    "github.com/yourusername/django-drf-go/core/settings"
)

func main() {
    // Load configuration
    conf := settings.MustLoad("config/settings.json")
    settings.Initialize(conf)
    
    // Validate
    if err := conf.Validate(); err != nil {
        log.Fatal(err)
    }
    
    // Use throughout app
    if settings.Get().Debug {
        log.Println("Running in debug mode")
    }
}
```

### 5.2 Accessing Settings

**Django:**
```python
from django.conf import settings

if settings.DEBUG:
    print("Debug mode")

page_size = settings.REST_FRAMEWORK.get('PAGE_SIZE', 20)
```

**Django-DRF-Go:**
```go
import "github.com/yourusername/django-drf-go/core/settings"

if settings.Get().Debug {
    fmt.Println("Debug mode")
}

pageSize := settings.Get().REST.Pagination.PageSize
```

### 5.3 Environment-Specific Settings

**Django:**
```python
# settings/base.py, settings/development.py, settings/production.py
from .base import *

DEBUG = False
ALLOWED_HOSTS = ['production.example.com']
```

**Django-DRF-Go:**
```
config/
  settings.json           # Base settings
  settings.development.json  # Dev overrides
  settings.production.json   # Prod overrides
```

```go
env := os.Getenv("APP_ENV") // "development" or "production"
conf := settings.LoadChain("config/settings", env)
```

---

## 6. Type-Safe Custom Settings

### 6.1 Extending Settings

```go
// myapp/settings.go
package myapp

// MyAppSettings for custom configuration
type MyAppSettings struct {
    FeatureFlags FeatureFlags `json:"feature_flags"`
    ExternalAPI  APIConfig    `json:"external_api"`
}

type FeatureFlags struct {
    EnableBeta    bool `json:"enable_beta"`
    MaxUploadSize int  `json:"max_upload_size"`
}

type APIConfig struct {
    BaseURL string `json:"base_url"`
    APIKey  string `json:"api_key" env:"EXTERNAL_API_KEY"`
    Timeout int    `json:"timeout" default:"30"`
}
```

### 6.2 Registering Custom Settings

```go
// Register custom settings type
func init() {
    settings.RegisterCustom("myapp", &MyAppSettings{})
}

// Access custom settings
func GetMyAppSettings() *MyAppSettings {
    return settings.GetCustom("myapp").(*MyAppSettings)
}
```

---

## 7. Settings Validation

### 7.1 Built-in Validators

```go
// Validation rules via struct tags
type Settings struct {
    SecretKey    string `required:"true" min_length:"32"`
    AllowedHosts []string `required_if:"Debug=false"`
    Port         int    `default:"8000" min:"1" max:"65535"`
}
```

### 7.2 Custom Validators

```go
// RegisterValidator adds custom validation
settings.RegisterValidator("secret_key", func(value string) error {
    if len(value) < 32 {
        return errors.New("SECRET_KEY must be at least 32 characters")
    }
    return nil
})
```

### 7.3 Startup Validation

```go
func (s *Settings) Validate() error {
    var errs []string
    
    // Required checks
    if s.SecretKey == "" {
        errs = append(errs, "SECRET_KEY is required")
    }
    
    // Production checks
    if !s.Debug {
        if len(s.AllowedHosts) == 0 {
            errs = append(errs, "ALLOWED_HOSTS required when DEBUG=false")
        }
    }
    
    // Database validation
    if db, ok := s.Databases["default"]; ok {
        if db.Engine == "" {
            errs = append(errs, "Database ENGINE is required")
        }
    } else {
        errs = append(errs, "Default database configuration required")
    }
    
    if len(errs) > 0 {
        return fmt.Errorf("settings validation failed:\n- %s", 
            strings.Join(errs, "\n- "))
    }
    
    return nil
}
```

---

## 8. Hot Reload (Development)

### 8.1 File Watching

```go
// EnableHotReload watches for config file changes (development only)
func EnableHotReload(configPath string, onChange func(*Settings)) error

// Usage
if settings.Get().Debug {
    settings.EnableHotReload("config/settings.json", func(s *Settings) {
        log.Println("Settings reloaded")
        settings.Initialize(s)
    })
}
```

### 8.2 Atomic Reload

```go
// Reload atomically replaces settings
func Reload(configPath string) error

// Thread-safe access during reload
conf := settings.Get() // Returns current settings atomically
```

---

## 9. Security Considerations

### 9.1 Sensitive Values

```go
// Mask sensitive fields in logs/errors
type DatabaseConfig struct {
    Password string `json:"password" sensitive:"true"`
}

// String() omits sensitive fields
func (s *Settings) String() string {
    // Returns settings with passwords masked
}
```

### 9.2 Secret Key Generation

```go
// GenerateSecretKey creates a cryptographically secure key
func GenerateSecretKey() string

// CLI usage: drf-manage generate-secret-key
```

---

## 10. Implementation Notes

### 10.1 Loading Priority

1. Default values from struct tags
2. Base configuration file
3. Environment-specific configuration file
4. Environment variables (highest priority)

### 10.2 Dependencies

| Package | Purpose |
|---------|---------|
| `encoding/json` | JSON parsing |
| `gopkg.in/yaml.v3` | YAML parsing |
| `os` | Environment variables |
| `sync/atomic` | Thread-safe config swapping |
| `fsnotify` | Optional hot reload |

---

## 11. Error Messages

| Error | Cause | Resolution |
|-------|-------|------------|
| `ErrConfigNotFound` | Config file doesn't exist | Check path or create file |
| `ErrInvalidFormat` | JSON/YAML parse error | Fix syntax errors |
| `ErrRequiredMissing` | Required setting not provided | Set in config or environment |
| `ErrInvalidValue` | Setting fails validation | Check constraints |
| `ErrEnvNotSet` | Required env var missing | Set environment variable |

---

## 12. Related PRDs

- [01_PRD_Core_Apps_System.md](./01_PRD_Core_Apps_System.md) - Apps use settings
- [03_PRD_Core_Middleware.md](./03_PRD_Core_Middleware.md) - Middleware chain from settings
- [07_PRD_ORM_Models.md](./07_PRD_ORM_Models.md) - Database config
- [17_PRD_DRF_Authentication.md](./17_PRD_DRF_Authentication.md) - Auth settings
