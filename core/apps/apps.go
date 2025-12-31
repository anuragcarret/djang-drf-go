// Package apps provides the application registry system.
// This is the foundation of django-drf-go, similar to django.apps.
package apps

import (
	"errors"
	"sync"
)

// Common errors
var (
	ErrDuplicateApp      = errors.New("app with this name already registered")
	ErrRegistryPopulated = errors.New("cannot register after Populate() has been called")
	ErrAppNotFound       = errors.New("app not found")
	ErrRegistryNotReady  = errors.New("registry not yet populated")
	ErrEmptyAppName      = errors.New("app name cannot be empty")
)

// App defines the contract for a django-drf-go application.
// Every application must implement this interface.
type App interface {
	// AppConfig returns the application's configuration
	AppConfig() *AppConfig

	// Ready is called when the app registry is fully populated.
	// Use this to perform initialization that requires other apps.
	Ready() error
}

// AppConfig holds application metadata and configuration.
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

	// DefaultAutoField specifies the default primary key type
	// Options: "BigAutoField", "AutoField", "UUIDField"
	DefaultAutoField string
}

// Registry manages installed applications.
type Registry struct {
	apps  map[string]App
	order []string // Maintains registration order
	ready bool
	mu    sync.RWMutex
}

// NewRegistry creates a new app registry.
func NewRegistry() *Registry {
	return &Registry{
		apps:  make(map[string]App),
		order: make([]string, 0),
	}
}

// Register adds an app to the registry.
// Must be called before Populate().
func (r *Registry) Register(app App) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.ready {
		return ErrRegistryPopulated
	}

	config := app.AppConfig()
	if config == nil || config.Name == "" {
		return ErrEmptyAppName
	}

	name := config.Name
	if _, exists := r.apps[name]; exists {
		return ErrDuplicateApp
	}

	r.apps[name] = app
	r.order = append(r.order, name)
	return nil
}

// Populate initializes all registered apps.
// Called once during application startup.
// Triggers Ready() on each app after all are registered.
func (r *Registry) Populate() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.ready {
		return nil // Already populated
	}

	// Call Ready() on each app in registration order
	for _, name := range r.order {
		app := r.apps[name]
		if err := app.Ready(); err != nil {
			return err
		}
	}

	r.ready = true
	return nil
}

// IsReady returns true if the registry has been populated.
func (r *Registry) IsReady() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.ready
}

// GetApp returns an app by name.
func (r *Registry) GetApp(name string) (App, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	app, exists := r.apps[name]
	if !exists {
		return nil, ErrAppNotFound
	}
	return app, nil
}

// GetApps returns all registered apps in registration order.
func (r *Registry) GetApps() []App {
	r.mu.RLock()
	defer r.mu.RUnlock()

	apps := make([]App, 0, len(r.order))
	for _, name := range r.order {
		apps = append(apps, r.apps[name])
	}
	return apps
}

// GetAppConfigs returns all app configurations.
func (r *Registry) GetAppConfigs() []*AppConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	configs := make([]*AppConfig, 0, len(r.order))
	for _, name := range r.order {
		configs = append(configs, r.apps[name].AppConfig())
	}
	return configs
}

// AppCount returns the number of registered apps.
func (r *Registry) AppCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.apps)
}

// Global registry instance
var Apps = NewRegistry()
