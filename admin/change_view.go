package admin

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/anuragcarret/djang-drf-go/core/apps"
	"github.com/anuragcarret/djang-drf-go/core/urls"
	"github.com/anuragcarret/djang-drf-go/orm/db"
	"github.com/anuragcarret/djang-drf-go/orm/queryset"
)

// ChangeView handles both GET (show pre-filled form) and POST (update record) for editing existing records
func (g *GenericAdmin[T]) ChangeView(w http.ResponseWriter, r *http.Request, database *db.DB) {
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

	// Extract ID from URL parameters (new router API)
	params := urls.GetParams(r)
	objectID, err := params.GetUint("id")
	if err != nil {
		http.Error(w, "Invalid ID in URL", http.StatusBadRequest)
		return
	}

	if r.Method == "POST" {
		// Handle form submission (update)
		g.handleChangePost(w, r, database, appName, modelName, objectID)
		return
	}

	// GET request - fetch record and render form
	g.renderChangeForm(w, r, database, appName, modelName, objectID, nil, nil)
}

func (g *GenericAdmin[T]) renderChangeForm(w http.ResponseWriter, r *http.Request, database *db.DB, appName, modelName string, objectID uint64, values map[string]interface{}, errors []string) {
	var zero T
	typ := reflect.TypeOf(zero)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	// Fetch the existing record
	qs := queryset.NewQuerySet[T](database)
	record, err := qs.GetByID(objectID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Record not found: %v", err), http.StatusNotFound)
		return
	}

	// Extract values from record if not provided (for initial render)
	if values == nil {
		values = g.extractValuesFromRecord(record, typ)
	}

	// Generate form fields with values
	fieldSets := g.generateFormFields(typ, values, nil)

	data := map[string]interface{}{
		"App":       appName,
		"ModelName": modelName,
		"FieldSets": fieldSets,
		"Errors":    errors,
		"ObjectID":  objectID, // Not nil for change view
		"Apps":      DefaultSite.getTemplateData()["Apps"],
	}

	DefaultSite.renderTemplate(w, "change_form.html", data)
}

func (g *GenericAdmin[T]) handleChangePost(w http.ResponseWriter, r *http.Request, database *db.DB, appName, modelName string, objectID uint64) {
	var zero T
	typ := reflect.TypeOf(zero)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Fetch existing record
	qs := queryset.NewQuerySet[T](database)
	record, err := qs.GetByID(objectID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Record not found: %v", err), http.StatusNotFound)
		return
	}

	// Get reflect value of the record
	recordValue := reflect.ValueOf(record)
	if recordValue.Kind() == reflect.Ptr {
		recordValue = recordValue.Elem()
	}

	values := make(map[string]interface{})
	var errors []string

	// Populate fields from form (updates the record)
	g.populateFieldsFromForm(recordValue, typ, r, &values, &errors)

	// Validate required fields
	g.validateRequiredFields(typ, r, &errors)

	if len(errors) > 0 {
		// Re-render form with errors
		g.renderChangeForm(w, r, database, appName, modelName, objectID, values, errors)
		return
	}

	// Update in database
	err = qs.Update(record)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Database error: %v", err))
		g.renderChangeForm(w, r, database, appName, modelName, objectID, values, errors)
		return
	}

	// Determine redirect based on button clicked
	if r.FormValue("_continue") != "" {
		// Save and continue editing
		http.Redirect(w, r, fmt.Sprintf("/admin/%s/%s/%d/change/", appName, modelName, objectID), http.StatusSeeOther)
	} else if r.FormValue("_addanother") != "" {
		// Save and add another
		http.Redirect(w, r, fmt.Sprintf("/admin/%s/%s/add/", appName, modelName), http.StatusSeeOther)
	} else {
		// Save and return to list
		http.Redirect(w, r, fmt.Sprintf("/admin/%s/%s/", appName, modelName), http.StatusSeeOther)
	}
}

// extractValuesFromRecord extracts field values from a record for form pre-population
func (g *GenericAdmin[T]) extractValuesFromRecord(record T, typ reflect.Type) map[string]interface{} {
	values := make(map[string]interface{})

	recordValue := reflect.ValueOf(record)
	if recordValue.Kind() == reflect.Ptr {
		recordValue = recordValue.Elem()
	}

	g.extractValuesRecursive(recordValue, typ, values)

	return values
}

func (g *GenericAdmin[T]) extractValuesRecursive(val reflect.Value, typ reflect.Type, values map[string]interface{}) {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// Skip auto-generated fields
		if g.shouldSkipField(field) {
			continue
		}

		// Handle embedded structs
		if field.Anonymous {
			embeddedVal := val.Field(i)
			g.extractValuesRecursive(embeddedVal, field.Type, values)
			continue
		}

		tag := field.Tag.Get("drf")
		if tag == "" {
			continue
		}

		fieldName := toSnakeCase(field.Name)
		fieldVal := val.Field(i)

		if !fieldVal.IsValid() || !fieldVal.CanInterface() {
			continue
		}

		// Convert value to string for form display
		var strValue string
		switch fieldVal.Kind() {
		case reflect.String:
			strValue = fieldVal.String()
		case reflect.Int, reflect.Int64, reflect.Int32:
			strValue = fmt.Sprintf("%d", fieldVal.Int())
		case reflect.Uint, reflect.Uint64:
			strValue = fmt.Sprintf("%d", fieldVal.Uint())
		case reflect.Bool:
			if fieldVal.Bool() {
				strValue = "true"
			} else {
				strValue = "false"
			}
		case reflect.Float64, reflect.Float32:
			strValue = fmt.Sprintf("%f", fieldVal.Float())
		case reflect.Struct:
			// Handle time.Time
			if t, ok := fieldVal.Interface().(interface{ Format(string) string }); ok {
				// For datetime-local: "2006-01-02T15:04"
				// For date: "2006-01-02"
				if strings.Contains(field.Tag.Get("drf"), "type=date") {
					strValue = t.Format("2006-01-02")
				} else {
					strValue = t.Format("2006-01-02T15:04")
				}
			}
		default:
			strValue = fmt.Sprintf("%v", fieldVal.Interface())
		}

		values[fieldName] = strValue
	}
}

// extractIDFromPath extracts the ID from a URL path like /admin/app/Model/123/change/
func extractIDFromPath(path string) (uint64, error) {
	// Pattern: /admin/{app}/{model}/{id}/change/
	// or: /{app}/{model}/{id}/change/ (if /admin prefix is stripped by router)

	// Match pattern: /{word}/{word}/{number}/change/
	re := regexp.MustCompile(`/([^/]+)/([^/]+)/(\d+)/change/?$`)
	matches := re.FindStringSubmatch(path)

	if len(matches) < 4 {
		return 0, fmt.Errorf("invalid path format: %s", path)
	}

	id, err := strconv.ParseUint(matches[3], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid ID: %s", matches[3])
	}

	return id, nil
}
