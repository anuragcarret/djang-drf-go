package routers

import (
	"fmt"
	"reflect"

	"github.com/anuragcarret/djang-drf-go/core/urls"
	"github.com/anuragcarret/djang-drf-go/drf/views"
)

// DefaultRouter automatically generates URL patterns for ViewSets
type DefaultRouter struct {
	registrations []registration
}

type registration struct {
	prefix  string
	viewset interface{}
}

func NewDefaultRouter() *DefaultRouter {
	return &DefaultRouter{
		registrations: make([]registration, 0),
	}
}

// Register adds a ViewSet to the router
func (r *DefaultRouter) Register(prefix string, viewset interface{}) {
	r.registrations = append(r.registrations, registration{prefix, viewset})
}

// URLs returns the generated URL patterns
func (r *DefaultRouter) URLs() []*urls.URLPattern {
	var patterns []*urls.URLPattern

	for _, reg := range r.registrations {
		prefix := reg.prefix
		if prefix != "" && prefix[len(prefix)-1] != '/' {
			prefix += "/"
		}

		// List endpoint (GET /prefix/, POST /prefix/)
		patterns = append(patterns, &urls.URLPattern{
			Pattern: prefix,
			Handler: views.Handler(reg.viewset),
			Name:    reg.prefix + "-list",
		})

		// Detail endpoint (GET /prefix/{id}/, PUT /prefix/{id}/, DELETE /prefix/{id}/)
		detailPattern := fmt.Sprintf("%s<int:id>/", prefix)
		patterns = append(patterns, &urls.URLPattern{
			Pattern: detailPattern,
			Handler: views.Handler(reg.viewset),
			Name:    reg.prefix + "-detail",
		})

		// TODO: Custom actions (@action) using reflection
		r.addCustomActions(&patterns, reg)
	}

	return patterns
}

func (r *DefaultRouter) addCustomActions(patterns *[]*urls.URLPattern, reg registration) {
	v := reflect.ValueOf(reg.viewset)
	method := v.MethodByName("Actions")
	if !method.IsValid() {
		return
	}

	// This assumes the ViewSet has an Actions() method returning []Action
	// For simplicity in this TDD step, we'll skip the complex reflection
	// until we define the Action struct in drf/views
}
