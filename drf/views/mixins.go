package views

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/anuragcarret/djang-drf-go/orm/db"
	"github.com/anuragcarret/djang-drf-go/orm/queryset"
)

// ListModelMixin provides a `List` method for listing a queryset
type ListModelMixin[T queryset.ModelInterface] struct {
	DB *db.DB
}

func (m *ListModelMixin[T]) List(c *Context) Response {
	qs := queryset.NewQuerySet[T](m.DB)
	qs = ApplyFilters(qs, c.Query)

	results, err := qs.All()
	if err != nil {
		return BadRequest(map[string]string{"error": err.Error()})
	}

	return OK(results)
}

// CreateModelMixin provides a `Create` method for creating a model instance
type CreateModelMixin[T queryset.ModelInterface] struct {
	DB *db.DB
}

func (m *CreateModelMixin[T]) Create(c *Context) Response {
	var instance T

	if err := c.Bind(&instance); err != nil {
		return BadRequest(map[string]string{"error": "Failed to bind data: " + err.Error()})
	}

	qs := queryset.NewQuerySet[T](m.DB)
	if err := qs.Create(instance); err != nil {
		return BadRequest(map[string]string{"error": "Failed to create: " + err.Error()})
	}

	return Created(instance)
}

// RetrieveModelMixin provides a `Retrieve` method for retrieving a single object
type RetrieveModelMixin[T queryset.ModelInterface] struct {
	DB          *db.DB
	LookupField string // Default: "id"
}

func (m *RetrieveModelMixin[T]) Retrieve(c *Context, lookupValue string) Response {
	if lookupValue == "" {
		return BadRequest(map[string]string{"error": "Lookup value is required"})
	}

	lookupField := m.LookupField
	if lookupField == "" {
		lookupField = "id"
	}

	qs := queryset.NewQuerySet[T](m.DB)

	// Convert lookup value to uint64 for ID field
	if lookupField == "id" {
		id, err := strconv.ParseUint(lookupValue, 10, 64)
		if err != nil {
			return BadRequest(map[string]string{"error": "Invalid ID format"})
		}

		obj, err := qs.GetByID(id)
		if err != nil {
			return NotFound("Object not found")
		}
		return OK(obj)
	}

	// For other fields, use filter
	results, err := qs.Filter(map[string]interface{}{lookupField: lookupValue}).All()
	if err != nil || len(results) == 0 {
		return NotFound("Object not found")
	}

	return OK(results[0])
}

// UpdateModelMixin provides `Update` and `PartialUpdate` methods
type UpdateModelMixin[T queryset.ModelInterface] struct {
	DB *db.DB
}

func (m *UpdateModelMixin[T]) Update(c *Context, id string) Response {
	return m.performUpdate(c, id, false)
}

func (m *UpdateModelMixin[T]) PartialUpdate(c *Context, id string) Response {
	return m.performUpdate(c, id, true)
}

func (m *UpdateModelMixin[T]) performUpdate(c *Context, id string, partial bool) Response {
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return BadRequest(map[string]string{"error": "Invalid ID format"})
	}

	qs := queryset.NewQuerySet[T](m.DB)

	// Get existing object
	existing, err := qs.GetByID(idUint)
	if err != nil {
		return NotFound("Object not found")
	}

	// Bind new data
	if err := c.Bind(existing); err != nil {
		return BadRequest(map[string]string{"error": "Failed to bind data: " + err.Error()})
	}

	// Update
	if err := qs.Update(existing); err != nil {
		return BadRequest(map[string]string{"error": "Failed to update: " + err.Error()})
	}

	return OK(existing)
}

// DestroyModelMixin provides a `Destroy` method for deleting an object
type DestroyModelMixin[T queryset.ModelInterface] struct {
	DB *db.DB
}

func (m *DestroyModelMixin[T]) Destroy(c *Context, id string) Response {
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return BadRequest(map[string]string{"error": "Invalid ID format"})
	}

	qs := queryset.NewQuerySet[T](m.DB)

	// Check if exists
	_, err = qs.GetByID(idUint)
	if err != nil {
		return NotFound("Object not found")
	}

	// Delete
	if err := qs.Delete(idUint); err != nil {
		return BadRequest(map[string]string{"error": "Failed to delete: " + err.Error()})
	}

	return NoContent()
}

// GenericAPIView is the base class for all generic views
// It provides common functionality like queryset access and object lookup
type GenericAPIView[T queryset.ModelInterface] struct {
	BaseAPIView
	DB          *db.DB
	LookupField string
}

func (v *GenericAPIView[T]) GetQueryset() *queryset.QuerySet[T] {
	return queryset.NewQuerySet[T](v.DB)
}

func (v *GenericAPIView[T]) GetObject(c *Context, lookupValue string) (T, error) {
	var zero T

	if lookupValue == "" {
		return zero, fmt.Errorf("lookup value is required")
	}

	lookupField := v.LookupField
	if lookupField == "" {
		lookupField = "id"
	}

	qs := v.GetQueryset()

	// Convert lookup value to uint64 for ID field
	if lookupField == "id" {
		id, err := strconv.ParseUint(lookupValue, 10, 64)
		if err != nil {
			return zero, fmt.Errorf("invalid ID format")
		}

		return qs.GetByID(id)
	}

	// For other fields, use filter
	results, err := qs.Filter(map[string]interface{}{lookupField: lookupValue}).All()
	if err != nil || len(results) == 0 {
		return zero, fmt.Errorf("object not found")
	}

	return results[0], nil
}

func (v *GenericAPIView[T]) FilterQueryset(qs *queryset.QuerySet[T], params url.Values) *queryset.QuerySet[T] {
	return ApplyFilters(qs, params)
}
