package admin

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"github.com/anuragcarret/djang-drf-go/core/apps"
	"github.com/anuragcarret/djang-drf-go/core/urls"
	"github.com/anuragcarret/djang-drf-go/orm/db"
	"github.com/anuragcarret/djang-drf-go/orm/queryset"
)

// AdminHandler defines the interface for a model's admin handler
type AdminHandler interface {
	RegisterRoutes(r *urls.Router, prefix string, db *db.DB)
	Config() *ModelAdmin
}

// ModelAdmin defines the interface for model administration customization
type ModelAdmin struct {
	ListDisplay  []string
	SearchFields []string
	ListFilter   []string
}

// AdminSite manages registered models and their admin configurations
type AdminSite struct {
	registry map[reflect.Type]AdminHandler
	mu       sync.RWMutex
}

// NewAdminSite creates a new AdminSite instance
func NewAdminSite() *AdminSite {
	return &AdminSite{
		registry: make(map[reflect.Type]AdminHandler),
	}
}

// Register (Legacy/Untyped) - Keeping for compatibility but delegating if possible
// Deprecated: Use Register[T] instead
func (s *AdminSite) Register(model interface{}, admin interface{}) error {
	return errors.New("use generic Register[T] instead")
}

// Register generic function to register a model with the admin site
func Register[T queryset.ModelInterface](s *AdminSite, config *ModelAdmin) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var zero T
	typ := reflect.TypeOf(zero)
	// If T is a pointer type, we want the element type for the key
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if _, exists := s.registry[typ]; exists {
		return errors.New("model already registered with admin")
	}

	s.registry[typ] = &GenericAdmin[T]{
		config: config,
	}
	return nil
}

// GenericAdmin implements AdminHandler for a specific type T
type GenericAdmin[T queryset.ModelInterface] struct {
	config *ModelAdmin
}

func (g *GenericAdmin[T]) Config() *ModelAdmin {
	return g.config
}

func (g *GenericAdmin[T]) RegisterRoutes(r *urls.Router, prefix string, database *db.DB) {
	// Register CRUD routes
	// /prefix/ -> List (GET), Create (POST)
	// /prefix/{id} -> Retrieve (GET), Update (PUT/PATCH), Delete (DELETE)

	// Delegate to generic views (implemented in views.go)
	r.Get(prefix, ListModelView[T](g.config, database), "admin_list")
}

// URLs returns a router with all admin routes
func (s *AdminSite) URLs(database *db.DB) *urls.Router {
	r := urls.NewRouter()

	// Admin Root
	r.Get("/", &AdminIndexView{Site: s, DB: database}, "admin_index")

	s.mu.RLock()
	defer s.mu.RUnlock()

	for typ, handler := range s.registry {
		// Create an instance to get app label
		// We need a pointer to the struct to pass to GetContainingApp if typ is struct
		var model interface{}
		if typ.Kind() == reflect.Struct {
			model = reflect.New(typ).Interface()
		} else {
			// Should not happen if strictly controlled, but handle just in case
			continue
		}

		appConfig := apps.Apps.GetContainingApp(model)
		if appConfig == nil {
			continue
		}

		appName := appConfig.Label
		modelName := typ.Name()

		// Route: /app/model
		routePath := fmt.Sprintf("/%s/%s", appName, modelName)
		handler.RegisterRoutes(r, routePath, database)
	}

	return r
}

type AdminIndexView struct {
	Site *AdminSite
	DB   *db.DB
}

func (v *AdminIndexView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Return list of apps and models
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Admin Index"}`))
}

// DefaultSite is the global default admin site instance
var DefaultSite = NewAdminSite()
