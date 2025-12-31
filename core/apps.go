package core

import (
	"fmt"
	"sync"
)

// AppConfig defines the interface for a django-drf-go app.
type AppConfig interface {
	GetName() string
	Ready() error
}

// Registry maintains the list of registered apps.
type Registry struct {
	apps map[string]AppConfig
	mu   sync.RWMutex
}

var globalRegistry = &Registry{
	apps: make(map[string]AppConfig),
}

// RegisterApp adds an app to the global registry.
func RegisterApp(config AppConfig) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.apps[config.GetName()] = config
}

// GetApp retrieves an app by name from the global registry.
func GetApp(name string) (AppConfig, bool) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	app, ok := globalRegistry.apps[name]
	return app, ok
}

// InitializeAllApps calls Ready() on all registered apps.
func InitializeAllApps() error {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	for name, app := range globalRegistry.apps {
		if err := app.Ready(); err != nil {
			return fmt.Errorf("failed to initialize app %s: %w", name, err)
		}
	}
	return nil
}

// ListApps returns names of all registered apps.
func ListApps() []string {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	var names []string
	for name := range globalRegistry.apps {
		names = append(names, name)
	}
	return names
}
