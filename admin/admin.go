package admin

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/django-drf-go/framework/orm"
)

// ModelAdmin defines how a model is displayed in the admin.
type ModelAdmin struct {
	Model        orm.Model
	ListDisplay  []string
	SearchFields []string
}

// AdminSite maintains the registration of models for the admin UI.
type AdminSite struct {
	models map[reflect.Type]*ModelAdmin
	mu     sync.RWMutex
}

var DefaultSite = &AdminSite{
	models: make(map[reflect.Type]*ModelAdmin),
}

// Register adds a model to the admin site.
func Register(model orm.Model, adminClass *ModelAdmin) {
	DefaultSite.mu.Lock()
	defer DefaultSite.mu.Unlock()

	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if adminClass == nil {
		adminClass = &ModelAdmin{Model: model}
	} else {
		adminClass.Model = model
	}

	DefaultSite.models[t] = adminClass
}

// GetRegisteredModels returns all registered models for the admin.
func GetRegisteredModels() []string {
	DefaultSite.mu.RLock()
	defer DefaultSite.mu.RUnlock()

	var names []string
	for t := range DefaultSite.models {
		names = append(names, t.Name())
	}
	return names
}

// RenderDashboard is a placeholder for the admin dashboard view.
func RenderDashboard() string {
	models := GetRegisteredModels()
	return fmt.Sprintf("Admin Dashboard - Registered Models: %v", models)
}
