# PRD: Core Apps System

> **Module:** `core/apps`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **Django Equivalent:** `django.apps`

---

## 1. Overview

### 1.1 Purpose

The Apps System is the foundation of Django-DRF-Go's modular architecture. It provides:

- **Application Registry**: Central catalog of all installed applications
- **Lifecycle Management**: Initialization, configuration, and readiness hooks
- **Dependency Resolution**: Ensures apps are loaded in correct order
- **Introspection**: Runtime discovery of models, routes, and signals

### 1.2 Design Philosophy

| Django Approach | Go-Idiomatic Approach |
|-----------------|----------------------|
| `AppConfig` class with metaclass magic | Interface-based `App` with explicit registration |
| Runtime class discovery | Compile-time interface satisfaction |
| String-based app references (`'myapp'`) | Type-safe app references via interfaces |
| `ready()` method override | `Ready()` interface method |

---

## 2. Core Concepts

### 2.1 App Interface

Every application must implement the `App` interface:

```go
// App defines the contract for a Django-DRF-Go application
type App interface {
    // AppConfig returns the application's configuration
    AppConfig() *AppConfig
    
    // Ready is called when the app registry is fully populated
    // Use this to perform initialization that requires other apps
    Ready() error
}
```

### 2.2 AppConfig Structure

```go
// AppConfig holds application metadata and configuration
type AppConfig struct {
    // Name is the unique identifier (e.g., "users", "orders")
    Name string
    
    // Label is the Python-style dotted path for compatibility
    // e.g., "myproject.users" - optional, defaults to Name
    Label string
    
    // VerboseName is human-readable name for admin/display
    VerboseName string
    
    // Path is the filesystem path to the app's root directory
    Path string
    
    // Models registered with this app (populated during registry)
    Models []ModelMeta
    
    // DefaultAutoField specifies the default primary key type
    // Options: "BigAutoField", "AutoField", "UUIDField"
    DefaultAutoField string
}
```

### 2.3 App Registry

The central registry manages all installed applications:

```go
// Registry manages installed applications
type Registry struct {
    // apps stores registered applications by name
    apps map[string]App
    
    // ready indicates if all apps have been initialized
    ready bool
    
    // models maps "app_label.model_name" to model metadata
    models map[string]ModelMeta
    
    // mu protects concurrent access
    mu sync.RWMutex
}

// Global registry instance
var Apps *Registry
```

---

## 3. API Specification

### 3.1 Registry Methods

#### Registration

```go
// Register adds an app to the registry
// Must be called before Populate()
func (r *Registry) Register(app App) error

// Example usage:
func init() {
    core.Apps.Register(&UserApp{})
}
```

#### Lifecycle

```go
// Populate initializes all registered apps
// Called once during application startup
// Triggers Ready() on each app after all are registered
func (r *Registry) Populate() error

// IsReady returns true if the registry has been populated
func (r *Registry) IsReady() bool
```

#### Lookups

```go
// GetApp returns an app by name
func (r *Registry) GetApp(name string) (App, error)

// GetApps returns all registered apps
func (r *Registry) GetApps() []App

// GetAppConfigs returns all app configurations
func (r *Registry) GetAppConfigs() []*AppConfig

// GetModel returns model metadata by "app.Model" format
func (r *Registry) GetModel(appLabel, modelName string) (ModelMeta, error)

// GetModels returns all models, optionally filtered by app
func (r *Registry) GetModels(appName ...string) []ModelMeta
```

### 3.2 App Implementation

#### Minimal App

```go
package users

import "github.com/yourusername/django-drf-go/core"

type UsersApp struct{}

func (a *UsersApp) AppConfig() *core.AppConfig {
    return &core.AppConfig{
        Name:        "users",
        VerboseName: "User Management",
    }
}

func (a *UsersApp) Ready() error {
    // Optional: Connect signals, validate configuration, etc.
    return nil
}

// Registration in init()
func init() {
    core.Apps.Register(&UsersApp{})
}
```

#### App with Dependencies

```go
package orders

import (
    "github.com/yourusername/django-drf-go/core"
    "myproject/users" // Imports bring in the app via init()
)

type OrdersApp struct{}

func (a *OrdersApp) AppConfig() *core.AppConfig {
    return &core.AppConfig{
        Name:        "orders",
        VerboseName: "Order Management",
    }
}

func (a *OrdersApp) Ready() error {
    // Safe to reference users app here - Ready() called after Populate()
    usersApp, _ := core.Apps.GetApp("users")
    
    // Connect to user signals
    users.UserCreated.Connect(a.onUserCreated)
    
    return nil
}

func (a *OrdersApp) onUserCreated(sender interface{}, user *users.User) {
    // Create default order settings for new user
}
```

---

## 4. Django Comparison

### 4.1 Defining an App

**Django (Python):**
```python
# users/apps.py
from django.apps import AppConfig

class UsersConfig(AppConfig):
    name = 'users'
    verbose_name = 'User Management'
    
    def ready(self):
        import users.signals  # noqa
```

**Django-DRF-Go:**
```go
// users/apps.go
package users

import "github.com/yourusername/django-drf-go/core"

type UsersApp struct{}

func (a *UsersApp) AppConfig() *core.AppConfig {
    return &core.AppConfig{
        Name:        "users",
        VerboseName: "User Management",
    }
}

func (a *UsersApp) Ready() error {
    // Import equivalent: Go imports in this file
    // Signals are connected here
    return nil
}

func init() {
    core.Apps.Register(&UsersApp{})
}
```

### 4.2 Registering Apps

**Django (settings.py):**
```python
INSTALLED_APPS = [
    'django.contrib.auth',
    'django.contrib.contenttypes',
    'users.apps.UsersConfig',
    'orders',
]
```

**Django-DRF-Go (main.go):**
```go
package main

import (
    "github.com/yourusername/django-drf-go/core"
    
    // Import apps - init() functions register them
    _ "github.com/yourusername/django-drf-go/contrib/auth"
    _ "github.com/yourusername/django-drf-go/contrib/contenttypes"
    _ "myproject/users"
    _ "myproject/orders"
)

func main() {
    // Populate the registry
    if err := core.Apps.Populate(); err != nil {
        log.Fatal(err)
    }
    
    // Start server...
}
```

### 4.3 Accessing Apps at Runtime

**Django:**
```python
from django.apps import apps

# Get specific app
users_config = apps.get_app_config('users')

# Get all models
all_models = apps.get_models()

# Get model by label
User = apps.get_model('users', 'User')
```

**Django-DRF-Go:**
```go
import "github.com/yourusername/django-drf-go/core"

// Get specific app
usersApp, err := core.Apps.GetApp("users")

// Get all models
allModels := core.Apps.GetModels()

// Get model by label
userMeta, err := core.Apps.GetModel("users", "User")
```

---

## 5. Lifecycle Flow

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Startup                       │
├─────────────────────────────────────────────────────────────┤
│  1. Import Phase                                             │
│     └─> Each app's init() calls Registry.Register()         │
│                                                              │
│  2. Populate Phase (core.Apps.Populate())                   │
│     ├─> Validate all app configurations                     │
│     ├─> Build dependency graph                              │
│     ├─> Register models with their apps                     │
│     └─> Mark registry as populating                         │
│                                                              │
│  3. Ready Phase                                              │
│     ├─> Call Ready() on each app in dependency order        │
│     ├─> Apps can now safely reference other apps            │
│     └─> Mark registry as ready                              │
│                                                              │
│  4. Runtime Phase                                            │
│     └─> Apps are immutable; queries via GetApp/GetModel     │
└─────────────────────────────────────────────────────────────┘
```

---

## 6. Error Handling

### 6.1 Registration Errors

| Error | Cause | Resolution |
|-------|-------|------------|
| `ErrDuplicateApp` | App with same name already registered | Use unique app names |
| `ErrRegistryPopulated` | Register() called after Populate() | Register all apps before Populate() |
| `ErrEmptyAppName` | AppConfig.Name is empty | Provide a non-empty Name |

### 6.2 Lookup Errors

| Error | Cause | Resolution |
|-------|-------|------------|
| `ErrAppNotFound` | App name doesn't exist | Check spelling or ensure app is registered |
| `ErrModelNotFound` | Model not registered with app | Register model in app's init or Ready |
| `ErrRegistryNotReady` | Query before Populate() completes | Call Populate() first |

---

## 7. Advanced Features

### 7.1 App Labels (Django Compatibility)

For migration compatibility and admin display:

```go
func (a *UsersApp) AppConfig() *core.AppConfig {
    return &core.AppConfig{
        Name:  "users",
        Label: "myproject.users",  // Python-style path
    }
}
```

### 7.2 Lazy Model References

For circular dependencies between models in different apps:

```go
// In orders app, referencing users.User
type Order struct {
    orm.Model
    
    // Lazy reference resolved after registry population
    User *orm.LazyForeignKey `drf:"to=users.User"`
}
```

### 7.3 App Hooks

Extended interface for additional lifecycle hooks:

```go
// AppWithMigrations provides migration hooks
type AppWithMigrations interface {
    App
    Migrations() []Migration
}

// AppWithURLs provides URL patterns
type AppWithURLs interface {
    App
    URLPatterns() []URLPattern
}

// AppWithAdmin provides admin configuration
type AppWithAdmin interface {
    App
    AdminConfig() *AdminConfig
}
```

---

## 8. Configuration Options

### 8.1 Registry Configuration

```go
// RegistryConfig controls registry behavior
type RegistryConfig struct {
    // StrictMode fails on any registration warning
    StrictMode bool
    
    // AllowLateRegistration permits Register() after Populate()
    // Use only for testing
    AllowLateRegistration bool
}

// Configure registry before population
core.Apps.Configure(RegistryConfig{
    StrictMode: true,
})
```

---

## 9. Testing Support

### 9.1 Test Isolation

```go
import "github.com/yourusername/django-drf-go/core/testing"

func TestMyApp(t *testing.T) {
    // Create isolated registry for testing
    registry := testing.NewTestRegistry()
    
    // Register only needed apps
    registry.Register(&MyApp{})
    registry.Populate()
    
    // Test...
}
```

### 9.2 Mock Apps

```go
// MockApp for testing
type MockApp struct {
    core.BaseApp
    ReadyCalled bool
}

func (m *MockApp) Ready() error {
    m.ReadyCalled = true
    return nil
}
```

---

## 10. Implementation Notes

### 10.1 Thread Safety

- Registry uses `sync.RWMutex` for concurrent read access
- Write operations (Register, Populate) are not concurrent-safe by design
- All registration must complete before serving requests

### 10.2 Performance Considerations

- App lookups are O(1) via map
- Model lookups use composite key: `"app_label.ModelName"`
- Registry is read-only after Populate() - no lock contention

### 10.3 Dependencies

| Package | Purpose |
|---------|---------|
| `sync` | Mutex for thread safety |
| `errors` | Error wrapping |
| `reflect` | Model introspection |

---

## 11. Future Enhancements

- [ ] Hot reloading for development (app re-registration)
- [ ] App versioning for migrations
- [ ] Pluggable app discovery (file-based scanning)
- [ ] GraphQL schema auto-generation from apps

---

## 12. Related PRDs

- [02_PRD_Core_Settings.md](./02_PRD_Core_Settings.md) - Settings accessed by AppConfig
- [03_PRD_Core_Middleware.md](./03_PRD_Core_Middleware.md) - Middleware uses app registry
- [05_PRD_Core_Signals.md](./05_PRD_Core_Signals.md) - Signals connected in Ready()
- [07_PRD_ORM_Models.md](./07_PRD_ORM_Models.md) - Models registered with apps
