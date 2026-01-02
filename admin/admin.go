package admin

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
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
	registry  map[reflect.Type]AdminHandler
	mu        sync.RWMutex
	Templates map[string]*template.Template
}

// NewAdminSite creates a new AdminSite instance
func NewAdminSite() *AdminSite {
	s := &AdminSite{
		registry:  make(map[reflect.Type]AdminHandler),
		Templates: make(map[string]*template.Template),
	}
	s.loadTemplates()
	return s
}

func (s *AdminSite) loadTemplates() {
	fsys, err := GetTemplateFS()
	if err != nil {
		log.Printf("Warning: could not get template FS: %v", err)
		return
	}

	baseContent, err := fs.ReadFile(fsys, "base.html")
	if err != nil {
		log.Printf("Warning: could not read base.html: %v", err)
		return
	}

	pages := []string{"index.html", "change_list.html"}
	for _, page := range pages {
		pageContent, err := fs.ReadFile(fsys, page)
		if err != nil {
			log.Printf("Warning: could not read %s: %v", page, err)
			continue
		}

		tpl := template.New(page).Funcs(template.FuncMap{
			"add": func(a, b int) int { return a + b },
		})
		_, err = tpl.New("base.html").Parse(string(baseContent))
		if err != nil {
			log.Printf("Error parsing base for %s: %v", page, err)
			continue
		}
		_, err = tpl.Parse(string(pageContent))
		if err != nil {
			log.Printf("Error parsing %s: %v", page, err)
			continue
		}
		s.Templates[page] = tpl
	}
}

func (s *AdminSite) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	s.mu.RLock()
	tpl, ok := s.Templates[name]
	s.mu.RUnlock()

	if !ok {
		http.Error(w, fmt.Sprintf("Template %s not found", name), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err := tpl.Execute(w, data)
	if err != nil {
		log.Printf("Error rendering template %s: %v", name, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	// API List View (JSON)
	r.Get("/api"+prefix+"/", ListModelView[T](g.config, database), "admin_api_list")

	// Template List View (HTML)
	r.Get(prefix+"/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g.ChangeListView(w, r, database)
	}), "admin_template_list")
}

// URLs returns a router with all admin routes
func (s *AdminSite) URLs(database *db.DB) *urls.Router {
	r := urls.NewRouter()

	// 1. API: Root
	r.Get("/api/", &AdminIndexView{Site: s, DB: database}, "admin_api_index")

	s.mu.RLock()
	defer s.mu.RUnlock()

	for typ, handler := range s.registry {
		var model interface{}
		if typ.Kind() == reflect.Struct {
			model = reflect.New(typ).Interface()
		} else {
			continue
		}

		appConfig := apps.Apps.GetContainingApp(model)
		if appConfig == nil {
			continue
		}

		appName := appConfig.Label
		modelName := typ.Name()

		// Route: /app/model/
		routePath := fmt.Sprintf("/%s/%s", appName, modelName)
		handler.RegisterRoutes(r, routePath, database)
	}

	// 2. Templates
	r.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.renderTemplate(w, "index.html", s.getTemplateData())
	}), "admin_root")

	return r
}

func (s *AdminSite) getTemplateData() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	type AppInfo struct {
		Name   string
		Models []struct {
			Name string
			App  string
		}
	}

	appMap := make(map[string]*AppInfo)

	for typ := range s.registry {
		var model interface{}
		if typ.Kind() == reflect.Struct {
			model = reflect.New(typ).Interface()
		} else {
			continue
		}

		appConfig := apps.Apps.GetContainingApp(model)
		if appConfig == nil {
			continue
		}

		label := appConfig.Label
		if _, ok := appMap[label]; !ok {
			appMap[label] = &AppInfo{Name: label}
		}
		appMap[label].Models = append(appMap[label].Models, struct {
			Name string
			App  string
		}{Name: typ.Name(), App: label})
	}

	appsList := make([]*AppInfo, 0, len(appMap))
	for _, app := range appMap {
		appsList = append(appsList, app)
	}

	return map[string]interface{}{
		"Apps": appsList,
	}
}
func (g *GenericAdmin[T]) ChangeListView(w http.ResponseWriter, r *http.Request, database *db.DB) {
	var zero T
	typ := reflect.TypeOf(zero)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	appConfig := apps.Apps.GetContainingApp(zero)
	appName := "Unknown"
	if appConfig != nil {
		appName = appConfig.Label
	}

	modelName := typ.Name()

	// 1. Columns
	columns := []string{"ID"}
	if g.config != nil && len(g.config.ListDisplay) > 0 {
		columns = g.config.ListDisplay
	} else {
		// Default to some fields if no ListDisplay
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			if !f.Anonymous && f.Tag.Get("drf") != "" && f.Name != "ID" {
				columns = append(columns, f.Name)
				if len(columns) > 5 {
					break
				}
			}
		}
	}

	// 2. Data
	qs := queryset.NewQuerySet[T](database)
	results, err := qs.All()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows := make([]map[string]interface{}, 0, len(results))

	for _, res := range results {
		val := reflect.ValueOf(res)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		rowValues := make([]interface{}, len(columns))
		for i, col := range columns {
			field, ok := findFieldByName(val, col)
			if ok {
				rowValues[i] = field.Interface()
			} else {
				rowValues[i] = "-"
			}
		}

		// Get ID for the link
		var rowID interface{}
		idField, ok := findFieldByName(val, "ID")
		if ok {
			rowID = idField.Interface()
		}

		rows = append(rows, map[string]interface{}{
			"ID":     rowID,
			"Values": rowValues,
		})
	}

	data := map[string]interface{}{
		"App":       appName,
		"ModelName": modelName,
		"Columns":   columns,
		"Rows":      rows,
		"Apps":      DefaultSite.getTemplateData()["Apps"],
	}

	DefaultSite.renderTemplate(w, "change_list.html", data)
}

func findFieldByName(v reflect.Value, name string) (reflect.Value, bool) {
	f := v.FieldByName(name)
	if f.IsValid() {
		return f, true
	}

	// Search in anonymous fields
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		sf := t.Field(i)
		if sf.Anonymous && sf.Type.Kind() == reflect.Struct {
			if subF, ok := findFieldByName(v.Field(i), name); ok {
				return subF, true
			}
		}
	}

	return reflect.Value{}, false
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
