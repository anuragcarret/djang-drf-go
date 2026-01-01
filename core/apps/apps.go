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

	// Models maps table names to model instances
	Models map[string]interface{}
}

// Registry manages installed applications.
type Registry struct {
	apps          map[string]App
	configs       map[string]*AppConfig  // Cache configs to persist state like Models map
	models        map[string]interface{} // Global mapping of table_name -> model
	pendingModels map[string][]pendingModel
	order         []string // Maintains registration order
	ready         bool
	mu            sync.RWMutex
}

type pendingModel struct {
	model     interface{}
	tableName string
}

// NewRegistry creates a new app registry.
func NewRegistry() *Registry {
	return &Registry{
		apps:          make(map[string]App),
		configs:       make(map[string]*AppConfig),
		models:        make(map[string]interface{}),
		pendingModels: make(map[string][]pendingModel),
		order:         make([]string, 0),
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
	r.configs[name] = config // Store the config instance
	r.order = append(r.order, name)

	// Process pending models for this app
	if pending, ok := r.pendingModels[name]; ok {
		if config.Models == nil {
			config.Models = make(map[string]interface{})
		}
		for _, m := range pending {
			config.Models[m.tableName] = m.model
			r.models[m.tableName] = m.model
		}
		delete(r.pendingModels, name)
	}
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

// RegisterModel adds a model to the app and the global registry.
func (r *Registry) RegisterModel(appLabel string, model interface{}, tableName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.apps[appLabel]; !exists {
		// Buffer it if app not registered yet (common during init())
		r.pendingModels[appLabel] = append(r.pendingModels[appLabel], pendingModel{
			model:     model,
			tableName: tableName,
		})
		return nil
	}

	config, ok := r.configs[appLabel]
	if !ok {
		// Should not happen if apps and configs are synced
		return ErrAppNotFound
	}

	if config.Models == nil {
		config.Models = make(map[string]interface{})
	}

	config.Models[tableName] = model
	r.models[tableName] = model
	return nil
}

// GetModel returns a model by its table name.
func (r *Registry) GetModel(tableName string) (interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	model, exists := r.models[tableName]
	if !exists {
		return nil, errors.New("model not found: " + tableName)
	}
	return model, nil
}

// GetAllModels returns all registered models across all apps.
func (r *Registry) GetAllModels() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to avoid mutation
	models := make(map[string]interface{})
	for k, v := range r.models {
		models[k] = v
	}
	return models
}

// AppCount returns the number of registered apps.
func (r *Registry) AppCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.apps)
}

// GetContainingApp returns the AppConfig for the app that contains the given model.
func (r *Registry) GetContainingApp(model interface{}) *AppConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 1. Get table name of the model
	type TableNamer interface {
		TableName() string
	}

	tn, ok := model.(TableNamer)
	if !ok {
		return nil
	}
	tableName := tn.TableName()

	// 2. Iterate configs to find the owner
	for _, config := range r.configs {
		if config.Models != nil {
			if _, ok := config.Models[tableName]; ok {
				return config
			}
		}
	}
	return nil
}

// Global registry instance
var Apps = NewRegistry()
