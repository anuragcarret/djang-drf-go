package admin

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/anuragcarret/djang-drf-go/core/apps"
	"github.com/anuragcarret/djang-drf-go/core/urls"
	"github.com/anuragcarret/djang-drf-go/orm/db"
	"github.com/anuragcarret/djang-drf-go/orm/queryset"
)

// DeleteView handles both GET (show confirmation) and POST (delete record)
func (g *GenericAdmin[T]) DeleteView(w http.ResponseWriter, r *http.Request, database *db.DB) {
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

	// Extract ID from URL parameters
	params := urls.GetParams(r)
	objectID, err := params.GetUint("id")
	if err != nil {
		http.Error(w, "Invalid ID in URL", http.StatusBadRequest)
		return
	}

	if r.Method == "POST" {
		// Handle deletion
		g.handleDeletePost(w, r, database, appName, modelName, objectID)
		return
	}

	// GET - show confirmation page
	g.renderDeleteConfirmation(w, r, database, appName, modelName, objectID)
}

func (g *GenericAdmin[T]) renderDeleteConfirmation(w http.ResponseWriter, r *http.Request, database *db.DB, appName, modelName string, objectID uint64) {
	// Fetch the object to verify it exists and get its representation
	qs := queryset.NewQuerySet[T](database)
	obj, err := qs.GetByID(objectID)
	if err != nil {
		http.Error(w, "Object not found", http.StatusNotFound)
		return
	}

	// Get string representation of object
	objRepr := getObjectRepr(obj)

	// TODO: Implement cascade detection
	// For now, we'll pass nil for related objects
	var relatedObjects []map[string]interface{}

	data := map[string]interface{}{
		"App":            appName,
		"ModelName":      modelName,
		"ObjectID":       objectID,
		"ObjectRepr":     objRepr,
		"RelatedObjects": relatedObjects,
		"Apps":           DefaultSite.getTemplateData()["Apps"],
	}

	DefaultSite.renderTemplate(w, "delete_confirmation.html", data)
}

func (g *GenericAdmin[T]) handleDeletePost(w http.ResponseWriter, r *http.Request, database *db.DB, appName, modelName string, objectID uint64) {
	qs := queryset.NewQuerySet[T](database)

	// Verify object exists before deletion
	_, err := qs.GetByID(objectID)
	if err != nil {
		http.Error(w, "Object not found", http.StatusNotFound)
		return
	}

	// Delete the object using its ID
	err = qs.Delete(objectID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Delete failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect to list view
	http.Redirect(w, r, fmt.Sprintf("/admin/%s/%s/", appName, modelName), http.StatusSeeOther)
}

// getObjectRepr returns a string representation of an object
func getObjectRepr(obj interface{}) string {
	// Try to use String() method if available
	if stringer, ok := obj.(interface{ String() string }); ok {
		return stringer.String()
	}

	// Try to get username, name, or title fields
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Try common field names
	for _, fieldName := range []string{"Username", "Name", "Title", "Email"} {
		field := val.FieldByName(fieldName)
		if field.IsValid() && field.Kind() == reflect.String {
			str := field.String()
			if str != "" {
				return str
			}
		}
	}

	// Fallback: use type name + ID
	idField := val.FieldByName("ID")
	if idField.IsValid() {
		return fmt.Sprintf("%s #%v", val.Type().Name(), idField.Interface())
	}

	return fmt.Sprintf("%s object", val.Type().Name())
}
