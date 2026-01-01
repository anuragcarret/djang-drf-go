package admin

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/anuragcarret/djang-drf-go/orm/db"
	"github.com/anuragcarret/djang-drf-go/orm/queryset"
)

// ListModelView returns a generic handler that lists records for model T
func ListModelView[T queryset.ModelInterface](config *ModelAdmin, database *db.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if database == nil {
			http.Error(w, "Database connection not available", http.StatusServiceUnavailable)
			return
		}
		qs := queryset.NewQuerySet[T](database)

		// For now, get all records
		// In future: Apply filters from r.URL.Query() based on config.ListFilter

		// We need a method to get results as a list.
		// QuerySet usually has methods like All() []T or similar.
		// Assuming All() returns ([]T, error)

		results, err := qs.All()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Build response
		var zero T
		response := map[string]interface{}{
			"action":  "list",
			"model":   reflect.TypeOf(zero).String(),
			"count":   len(results),
			"results": results,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
}
