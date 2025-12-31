package admin

import (
	"fmt"

	"github.com/anuragcarret/djang-drf-go/core/urls"
)

// ModelInterface mirrors orm.ModelInterface
type ModelInterface interface {
	TableName() string
}

// AdminInterface defines how a model is displayed in the admin
type AdminInterface interface {
	GetListDisplay() []string
	GetListFilter() []string
	GetSearchFields() []string
}

// ModelAdmin provides a default implementation of AdminInterface
type ModelAdmin struct {
	ListDisplay  []string
	ListFilter   []string
	SearchFields []string
	Ordering     []string
}

func (a *ModelAdmin) GetListDisplay() []string  { return a.ListDisplay }
func (a *ModelAdmin) GetListFilter() []string   { return a.ListFilter }
func (a *ModelAdmin) GetSearchFields() []string { return a.SearchFields }

// AdminSite manages the admin interface
type AdminSite struct {
	Name      string
	URLPrefix string
	Models    map[string]AdminInterface
}

func NewAdminSite(name, prefix string) *AdminSite {
	return &AdminSite{
		Name:      name,
		URLPrefix: prefix,
		Models:    make(map[string]AdminInterface),
	}
}

// Register adds a model to the admin site
func (s *AdminSite) Register(model ModelInterface, admin AdminInterface) {
	if admin == nil {
		admin = &ModelAdmin{}
	}
	s.Models[model.TableName()] = admin
}

// URLs generates the admin URL patterns
func (s *AdminSite) URLs() []*urls.URLPattern {
	var patterns []*urls.URLPattern

	for tableName := range s.Models {
		// List view
		patterns = append(patterns, &urls.URLPattern{
			Pattern: tableName + "/",
			Name:    fmt.Sprintf("admin-%s-list", tableName),
		})

		// Add view
		patterns = append(patterns, &urls.URLPattern{
			Pattern: fmt.Sprintf("%s/add/", tableName),
			Name:    fmt.Sprintf("admin-%s-add", tableName),
		})

		// Change view
		patterns = append(patterns, &urls.URLPattern{
			Pattern: fmt.Sprintf("%s/<int:id>/", tableName),
			Name:    fmt.Sprintf("admin-%s-change", tableName),
		})
	}

	return patterns
}

var Site = NewAdminSite("admin", "/admin")
