package filters

import (
	"net/url"
	"strings"

	"github.com/anuragcarret/djang-drf-go/orm/queryset"
)

// FilterBackend defines the interface for filter backends
type FilterBackend interface {
	FilterQueryset(qs interface{}, params url.Values) interface{}
}

// DjangoFilterBackend implements field-based filtering
type DjangoFilterBackend struct {
	FilterFields []string // Allowed filter fields
}

func NewDjangoFilterBackend(fields []string) *DjangoFilterBackend {
	return &DjangoFilterBackend{
		FilterFields: fields,
	}
}

func (f *DjangoFilterBackend) FilterQueryset(qs interface{}, params url.Values) interface{} {
	// Type assertion to QuerySet
	// This is generic, so we'll use reflection or interface methods

	filters := make(queryset.Q)

	for _, field := range f.FilterFields {
		// Check for base field
		if val := params.Get(field); val != "" {
			filters[field] = val
		}

		// Check for lookup expressions
		lookups := []string{"gt", "gte", "lt", "lte", "contains", "icontains", "in", "iexact"}
		for _, lookup := range lookups {
			key := field + "__" + lookup
			if val := params.Get(key); val != "" {
				// Handle 'in' lookup specially (comma-separated values)
				if lookup == "in" {
					filters[key] = strings.Split(val, ",")
				} else {
					filters[key] = val
				}
			}
		}
	}

	// Apply filters to queryset
	// This would call qs.Filter(filters) if we have the right type
	// For now, return as-is (implementation depends on QuerySet interface)
	return qs
}

// SearchFilter implements text search across multiple fields
type SearchFilter struct {
	SearchFields []string // Fields to search in
	SearchParam  string   // Query parameter name (default: "search")
}

func NewSearchFilter(fields []string) *SearchFilter {
	return &SearchFilter{
		SearchFields: fields,
		SearchParam:  "search",
	}
}

func (f *SearchFilter) FilterQueryset(qs interface{}, params url.Values) interface{} {
	searchTerm := params.Get(f.SearchParam)
	if searchTerm == "" {
		return qs
	}

	// Sanitize search term
	searchTerm = strings.TrimSpace(searchTerm)
	if searchTerm == "" {
		return qs
	}

	// Build OR filters for all search fields
	// In a real implementation, this would use QuerySet's Q objects for OR queries
	// For now, we'll build filters for each field with icontains

	filters := make(queryset.Q)
	for _, field := range f.SearchFields {
		key := field + "__icontains"
		filters[key] = searchTerm
	}

	return qs
}

// OrderingFilter implements result ordering
type OrderingFilter struct {
	OrderingFields  []string // Allowed ordering fields
	OrderingParam   string   // Query parameter name (default: "ordering")
	DefaultOrdering []string // Default ordering if not specified
}

func NewOrderingFilter(fields []string) *OrderingFilter {
	return &OrderingFilter{
		OrderingFields:  fields,
		OrderingParam:   "ordering",
		DefaultOrdering: []string{},
	}
}

func (f *OrderingFilter) FilterQueryset(qs interface{}, params url.Values) interface{} {
	orderingStr := params.Get(f.OrderingParam)

	var ordering []string
	if orderingStr != "" {
		// Parse comma-separated ordering fields
		ordering = strings.Split(orderingStr, ",")
	} else {
		ordering = f.DefaultOrdering
	}

	if len(ordering) == 0 {
		return qs
	}

	// Validate ordering fields against whitelist
	validOrdering := []string{}
	for _, field := range ordering {
		// Remove leading - for descending
		cleanField := strings.TrimPrefix(field, "-")

		// Check if field is in allowed list
		isAllowed := false
		for _, allowed := range f.OrderingFields {
			if cleanField == allowed {
				isAllowed = true
				break
			}
		}

		if isAllowed {
			validOrdering = append(validOrdering, field)
		}
	}

	if len(validOrdering) == 0 {
		return qs
	}

	// Apply ordering to queryset
	// This would call qs.OrderBy(validOrdering...) if we have the right type
	return qs
}

// CombineFilters applies multiple filter backends in sequence
func CombineFilters(qs interface{}, params url.Values, backends []FilterBackend) interface{} {
	result := qs
	for _, backend := range backends {
		result = backend.FilterQueryset(result, params)
	}
	return result
}

// Helper function to extract filter params from URL
func ExtractFilterParams(params url.Values) map[string]interface{} {
	filters := make(map[string]interface{})
	for key, values := range params {
		if len(values) > 0 {
			// Skip pagination params
			if key == "page" || key == "page_size" || key == "limit" || key == "offset" || key == "cursor" {
				continue
			}
			filters[key] = values[0]
		}
	}
	return filters
}
