package admin

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/anuragcarret/djang-drf-go/core/apps"
	"github.com/anuragcarret/djang-drf-go/orm/db"
	"github.com/anuragcarret/djang-drf-go/orm/queryset"
)

// AddView handles both GET (show form) and POST (create record) for adding new records
func (g *GenericAdmin[T]) AddView(w http.ResponseWriter, r *http.Request, database *db.DB) {
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

	if r.Method == "POST" {
		// Handle form submission
		g.handleAddPost(w, r, database, appName, modelName)
		return
	}

	// GET request - render form
	g.renderAddForm(w, r, appName, modelName, nil, nil)
}

func (g *GenericAdmin[T]) renderAddForm(w http.ResponseWriter, r *http.Request, appName, modelName string, values map[string]interface{}, errors []string) {
	var zero T
	typ := reflect.TypeOf(zero)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	// Generate form fields
	fieldSets := g.generateFormFields(typ, values, nil)

	data := map[string]interface{}{
		"App":       appName,
		"ModelName": modelName,
		"FieldSets": fieldSets,
		"Errors":    errors,
		"ObjectID":  nil, // nil for add view
		"Apps":      DefaultSite.getTemplateData()["Apps"],
	}

	DefaultSite.renderTemplate(w, "change_form.html", data)
}

func (g *GenericAdmin[T]) handleAddPost(w http.ResponseWriter, r *http.Request, database *db.DB, appName, modelName string) {
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

	// Create new instance
	instance := reflect.New(typ)
	values := make(map[string]interface{})
	var errors []string

	// Populate fields from form - use a helper to avoid duplicates
	g.populateFieldsFromForm(instance.Elem(), typ, r, &values, &errors)

	// Validate required fields
	g.validateRequiredFields(typ, r, &errors)

	if len(errors) > 0 {
		// Re-render form with errors
		g.renderAddForm(w, r, appName, modelName, values, errors)
		return
	}

	// Save to database
	qs := queryset.NewQuerySet[T](database)
	err := qs.Create(instance.Interface().(T))
	if err != nil {
		errors = append(errors, fmt.Sprintf("Database error: %v", err))
		g.renderAddForm(w, r, appName, modelName, values, errors)
		return
	}

	// Get the ID of the created object
	idField := instance.Elem().FieldByName("ID")
	var objectID uint64
	if idField.IsValid() {
		objectID = idField.Uint()
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

func (g *GenericAdmin[T]) generateFormFields(typ reflect.Type, values map[string]interface{}, fieldErrors map[string]string) []FormFieldSet {
	// For now, create a single fieldset with all fields
	fields := make([]FormField, 0)

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// Skip auto-generated fields
		if g.shouldSkipField(field) {
			continue
		}

		tag := field.Tag.Get("drf")
		if tag == "" && !field.Anonymous {
			continue
		}

		// Handle embedded structs
		if field.Anonymous {
			embeddedFields := g.generateFormFields(field.Type, values, fieldErrors)
			for _, fs := range embeddedFields {
				fields = append(fields, fs.Fields...)
			}
			continue
		}

		fieldName := toSnakeCase(field.Name)

		formField := FormField{
			Name:     fieldName,
			Label:    field.Name,
			Required: !strings.Contains(tag, "null") && !strings.Contains(tag, "blank"),
			ReadOnly: strings.Contains(tag, "readonly"),
		}

		// Set value from form values if available
		if values != nil {
			if val, ok := values[fieldName]; ok {
				formField.Value = val
			}
		}

		// Set error if available
		if fieldErrors != nil {
			if err, ok := fieldErrors[fieldName]; ok {
				formField.Error = err
			}
		}

		// Determine widget type based on Go type
		switch field.Type.Kind() {
		case reflect.String:
			// Check for email field
			if strings.Contains(strings.ToLower(field.Name), "email") {
				formField.Widget = "email"
			} else if strings.Contains(tag, "max_length") {
				formField.Widget = "text"
				// Extract max_length
				parts := strings.Split(tag, ";")
				for _, part := range parts {
					if strings.HasPrefix(part, "max_length=") {
						if ml, err := strconv.Atoi(strings.TrimPrefix(part, "max_length=")); err == nil {
							formField.MaxLength = ml
						}
					}
				}
			} else {
				formField.Widget = "textarea"
			}

		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Uint, reflect.Uint64:
			formField.Widget = "number"

		case reflect.Float64, reflect.Float32:
			formField.Widget = "number"
			formField.Step = "0.01"

		case reflect.Bool:
			formField.Widget = "checkbox"

		case reflect.Struct:
			if field.Type == reflect.TypeOf(time.Time{}) {
				// Check if it's a date or datetime field
				if strings.Contains(tag, "type=date") {
					formField.Widget = "date"
				} else {
					formField.Widget = "datetime"
				}
			}

		default:
			formField.Widget = "text"
		}

		fields = append(fields, formField)
	}

	return []FormFieldSet{
		{
			Name:   "",
			Fields: fields,
		},
	}
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// populateFieldsFromForm recursively populates struct fields from form data
// This handles embedded structs properly to avoid duplicate columns
func (g *GenericAdmin[T]) populateFieldsFromForm(val reflect.Value, typ reflect.Type, r *http.Request, values *map[string]interface{}, errors *[]string) {
	processedFields := make(map[string]bool)

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// Skip auto-generated fields
		if g.shouldSkipField(field) {
			continue
		}

		// Handle embedded structs recursively
		if field.Anonymous {
			embeddedVal := val.Field(i)
			g.populateFieldsFromForm(embeddedVal, field.Type, r, values, errors)
			continue
		}

		tag := field.Tag.Get("drf")
		if tag == "" {
			continue
		}

		fieldName := toSnakeCase(field.Name)

		// Skip if we've already processed this field (from embedded struct)
		if processedFields[fieldName] {
			continue
		}
		processedFields[fieldName] = true

		formValue := r.FormValue(fieldName)
		(*values)[fieldName] = formValue

		// Set field value based on type
		fieldVal := val.Field(i)
		if !fieldVal.CanSet() {
			continue
		}

		switch field.Type.Kind() {
		case reflect.String:
			fieldVal.SetString(formValue)

		case reflect.Int, reflect.Int64, reflect.Int32:
			if formValue != "" {
				if intVal, err := strconv.ParseInt(formValue, 10, 64); err == nil {
					fieldVal.SetInt(intVal)
				} else {
					*errors = append(*errors, fmt.Sprintf("%s: invalid integer value", field.Name))
				}
			}

		case reflect.Uint, reflect.Uint64, reflect.Uint32:
			if formValue != "" {
				if uintVal, err := strconv.ParseUint(formValue, 10, 64); err == nil {
					fieldVal.SetUint(uintVal)
				} else {
					*errors = append(*errors, fmt.Sprintf("%s: invalid integer value", field.Name))
				}
			}

		case reflect.Bool:
			// Checkboxes: present = true, absent = false
			fieldVal.SetBool(formValue == "true" || formValue == "on" || formValue == "1")

		case reflect.Float64, reflect.Float32:
			if formValue != "" {
				if floatVal, err := strconv.ParseFloat(formValue, 64); err == nil {
					fieldVal.SetFloat(floatVal)
				} else {
					*errors = append(*errors, fmt.Sprintf("%s: invalid number value", field.Name))
				}
			}

		case reflect.Struct:
			if field.Type == reflect.TypeOf(time.Time{}) {
				if formValue != "" {
					// Try parsing different time formats
					var t time.Time
					var parseErr error

					// Try date format first
					t, parseErr = time.Parse("2006-01-02", formValue)
					if parseErr != nil {
						// Try datetime-local format
						t, parseErr = time.Parse("2006-01-02T15:04", formValue)
					}
					if parseErr != nil {
						// Try full datetime format
						t, parseErr = time.Parse(time.RFC3339, formValue)
					}

					if parseErr == nil {
						fieldVal.Set(reflect.ValueOf(t))
					} else {
						*errors = append(*errors, fmt.Sprintf("%s: invalid date/time format", field.Name))
					}
				}
			}
		}
	}
}

// shouldSkipField determines if a field should be skipped during form processing
func (g *GenericAdmin[T]) shouldSkipField(field reflect.StructField) bool {
	// Skip ID and Model fields
	if field.Name == "ID" || field.Name == "Model" {
		return true
	}

	// Skip auto-generated timestamp fields
	if field.Name == "CreatedAt" || field.Name == "UpdatedAt" || field.Name == "DeletedAt" {
		return true
	}

	tag := field.Tag.Get("drf")
	// Skip fields with auto_ tags (auto_now, auto_now_add)
	if strings.Contains(tag, "auto_now") || strings.Contains(tag, "auto_now_add") {
		return true
	}

	return false
}

// validateRequiredFields validates that all required fields have values
func (g *GenericAdmin[T]) validateRequiredFields(typ reflect.Type, r *http.Request, errors *[]string) {
	g.validateRequiredFieldsRecursive(typ, r, errors, make(map[string]bool))
}

func (g *GenericAdmin[T]) validateRequiredFieldsRecursive(typ reflect.Type, r *http.Request, errors *[]string, processed map[string]bool) {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// Handle embedded structs
		if field.Anonymous {
			g.validateRequiredFieldsRecursive(field.Type, r, errors, processed)
			continue
		}

		// Skip auto-generated fields
		if g.shouldSkipField(field) {
			continue
		}

		tag := field.Tag.Get("drf")
		if tag == "" {
			continue
		}

		fieldName := toSnakeCase(field.Name)

		// Skip if already validated
		if processed[fieldName] {
			continue
		}
		processed[fieldName] = true

		// Check if field is required (has no "null" or "blank" option)
		if !strings.Contains(tag, "null") && !strings.Contains(tag, "blank") {
			formValue := r.FormValue(fieldName)
			if formValue == "" && field.Type.Kind() != reflect.Bool {
				*errors = append(*errors, fmt.Sprintf("%s is required", field.Name))
			}
		}
	}
}
