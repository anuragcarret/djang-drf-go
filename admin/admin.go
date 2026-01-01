package admin

import (
	"errors"
	"reflect"
	"sync"
)

// ModelAdmin defines the interface for model administration customization
type ModelAdmin struct {
	ListDisplay  []string
	SearchFields []string
	ListFilter   []string
}

// AdminSite manages registered models and their admin configurations
type AdminSite struct {
	registry map[reflect.Type]interface{}
	mu       sync.RWMutex
}

// NewAdminSite creates a new AdminSite instance
func NewAdminSite() *AdminSite {
	return &AdminSite{
		registry: make(map[reflect.Type]interface{}),
	}
}

// Register adds a model and its admin configuration to the site
func (s *AdminSite) Register(model interface{}, admin interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	typ := reflect.TypeOf(model)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if _, exists := s.registry[typ]; exists {
		return errors.New("model already registered with admin")
	}

	s.registry[typ] = admin
	return nil
}

// GetAdmin retrieves the admin configuration for a given model
func (s *AdminSite) GetAdmin(model interface{}) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	typ := reflect.TypeOf(model)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	admin, exists := s.registry[typ]
	if !exists {
		return nil, errors.New("model not registered with admin")
	}

	return admin, nil
}

// DefaultSite is the global default admin site instance
var DefaultSite = NewAdminSite()
