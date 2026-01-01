package views

import (
	"github.com/anuragcarret/djang-drf-go/orm/queryset"
)

// FilterBackend defines the interface for filtering querysets
type FilterBackend interface {
	FilterQuerySet(c *Context, qs interface{}) interface{}
}

// DjangoFilterBackend implements Django-style filtering using __ lookups
type DjangoFilterBackend struct{}

func (f *DjangoFilterBackend) FilterQuerySet(c *Context, q interface{}) interface{} {
	// We use reflection/type-assertion to handle the generic QuerySet
	// This is a simplified version that works with our current architecture

	// Since we can't easily type-assert generic types without knowing T,
	// we'll assume the view passes a compatible object or we use a more generic interface.
	return q
}

// ApplyFilters applies query parameters directly to a QuerySet
func ApplyFilters[T queryset.ModelInterface](q *queryset.QuerySet[T], queryParams map[string][]string) *queryset.QuerySet[T] {
	newQ := q
	for key, values := range queryParams {
		if len(values) > 0 {
			// Skip framework parameters like 'page'
			if key == "page" || key == "page_size" {
				continue
			}
			newQ = newQ.Filter(queryset.Q{key: values[0]})
		}
	}
	return newQ
}
