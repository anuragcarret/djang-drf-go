package views

import "github.com/anuragcarret/djang-drf-go/orm/queryset"

// ListAPIView - Concrete view for listing a queryset
type ListAPIView[T queryset.ModelInterface] struct {
	GenericAPIView[T]
	ListModelMixin[T]
}

func (v *ListAPIView[T]) Get(c *Context) Response {
	v.ListModelMixin.DB = v.GenericAPIView.DB
	return v.ListModelMixin.List(c)
}

// CreateAPIView - Concrete view for creating a model instance
type CreateAPIView[T queryset.ModelInterface] struct {
	GenericAPIView[T]
	CreateModelMixin[T]
}

func (v *CreateAPIView[T]) Post(c *Context) Response {
	v.CreateModelMixin.DB = v.GenericAPIView.DB
	return v.CreateModelMixin.Create(c)
}

// RetrieveAPIView - Concrete view for retrieving a single object
type RetrieveAPIView[T queryset.ModelInterface] struct {
	GenericAPIView[T]
	RetrieveModelMixin[T]
}

func (v *RetrieveAPIView[T]) Get(c *Context) Response {
	v.RetrieveModelMixin.DB = v.GenericAPIView.DB
	v.RetrieveModelMixin.LookupField = v.GenericAPIView.LookupField

	// Extract ID from URL (assuming it's passed as a query param or path param)
	id := c.Query.Get("id")
	if id == "" {
		// Try to get from URL path params if available
		id = c.URLParam("id")
	}

	return v.RetrieveModelMixin.Retrieve(c, id)
}

// UpdateAPIView - Concrete view for updating an object
type UpdateAPIView[T queryset.ModelInterface] struct {
	GenericAPIView[T]
	UpdateModelMixin[T]
}

func (v *UpdateAPIView[T]) Put(c *Context) Response {
	v.UpdateModelMixin.DB = v.GenericAPIView.DB
	id := c.Query.Get("id")
	if id == "" {
		id = c.URLParam("id")
	}
	return v.UpdateModelMixin.Update(c, id)
}

func (v *UpdateAPIView[T]) Patch(c *Context) Response {
	v.UpdateModelMixin.DB = v.GenericAPIView.DB
	id := c.Query.Get("id")
	if id == "" {
		id = c.URLParam("id")
	}
	return v.UpdateModelMixin.PartialUpdate(c, id)
}

// DestroyAPIView - Concrete view for deleting an object
type DestroyAPIView[T queryset.ModelInterface] struct {
	GenericAPIView[T]
	DestroyModelMixin[T]
}

func (v *DestroyAPIView[T]) Delete(c *Context) Response {
	v.DestroyModelMixin.DB = v.GenericAPIView.DB
	id := c.Query.Get("id")
	if id == "" {
		id = c.URLParam("id")
	}
	return v.DestroyModelMixin.Destroy(c, id)
}

// ListCreateAPIView - Concrete view for list and create operations
type ListCreateAPIView[T queryset.ModelInterface] struct {
	GenericAPIView[T]
	ListModelMixin[T]
	CreateModelMixin[T]
}

func (v *ListCreateAPIView[T]) Get(c *Context) Response {
	v.ListModelMixin.DB = v.GenericAPIView.DB
	return v.ListModelMixin.List(c)
}

func (v *ListCreateAPIView[T]) Post(c *Context) Response {
	v.CreateModelMixin.DB = v.GenericAPIView.DB
	return v.CreateModelMixin.Create(c)
}

// RetrieveUpdateAPIView - Concrete view for retrieve and update operations
type RetrieveUpdateAPIView[T queryset.ModelInterface] struct {
	GenericAPIView[T]
	RetrieveModelMixin[T]
	UpdateModelMixin[T]
}

func (v *RetrieveUpdateAPIView[T]) Get(c *Context) Response {
	v.RetrieveModelMixin.DB = v.GenericAPIView.DB
	v.RetrieveModelMixin.LookupField = v.GenericAPIView.LookupField
	id := c.Query.Get("id")
	if id == "" {
		id = c.URLParam("id")
	}
	return v.RetrieveModelMixin.Retrieve(c, id)
}

func (v *RetrieveUpdateAPIView[T]) Put(c *Context) Response {
	v.UpdateModelMixin.DB = v.GenericAPIView.DB
	id := c.Query.Get("id")
	if id == "" {
		id = c.URLParam("id")
	}
	return v.UpdateModelMixin.Update(c, id)
}

func (v *RetrieveUpdateAPIView[T]) Patch(c *Context) Response {
	v.UpdateModelMixin.DB = v.GenericAPIView.DB
	id := c.Query.Get("id")
	if id == "" {
		id = c.URLParam("id")
	}
	return v.UpdateModelMixin.PartialUpdate(c, id)
}

// RetrieveDestroyAPIView - Concrete view for retrieve and delete operations
type RetrieveDestroyAPIView[T queryset.ModelInterface] struct {
	GenericAPIView[T]
	RetrieveModelMixin[T]
	DestroyModelMixin[T]
}

func (v *RetrieveDestroyAPIView[T]) Get(c *Context) Response {
	v.RetrieveModelMixin.DB = v.GenericAPIView.DB
	v.RetrieveModelMixin.LookupField = v.GenericAPIView.LookupField
	id := c.Query.Get("id")
	if id == "" {
		id = c.URLParam("id")
	}
	return v.RetrieveModelMixin.Retrieve(c, id)
}

func (v *RetrieveDestroyAPIView[T]) Delete(c *Context) Response {
	v.DestroyModelMixin.DB = v.GenericAPIView.DB
	id := c.Query.Get("id")
	if id == "" {
		id = c.URLParam("id")
	}
	return v.DestroyModelMixin.Destroy(c, id)
}

// RetrieveUpdateDestroyAPIView - Full CRUD view (except Create and List)
type RetrieveUpdateDestroyAPIView[T queryset.ModelInterface] struct {
	GenericAPIView[T]
	RetrieveModelMixin[T]
	UpdateModelMixin[T]
	DestroyModelMixin[T]
}

func (v *RetrieveUpdateDestroyAPIView[T]) Get(c *Context) Response {
	v.RetrieveModelMixin.DB = v.GenericAPIView.DB
	v.RetrieveModelMixin.LookupField = v.GenericAPIView.LookupField
	id := c.Query.Get("id")
	if id == "" {
		id = c.URLParam("id")
	}
	return v.RetrieveModelMixin.Retrieve(c, id)
}

func (v *RetrieveUpdateDestroyAPIView[T]) Put(c *Context) Response {
	v.UpdateModelMixin.DB = v.GenericAPIView.DB
	id := c.Query.Get("id")
	if id == "" {
		id = c.URLParam("id")
	}
	return v.UpdateModelMixin.Update(c, id)
}

func (v *RetrieveUpdateDestroyAPIView[T]) Patch(c *Context) Response {
	v.UpdateModelMixin.DB = v.GenericAPIView.DB
	id := c.Query.Get("id")
	if id == "" {
		id = c.URLParam("id")
	}
	return v.UpdateModelMixin.PartialUpdate(c, id)
}

func (v *RetrieveUpdateDestroyAPIView[T]) Delete(c *Context) Response {
	v.DestroyModelMixin.DB = v.GenericAPIView.DB
	id := c.Query.Get("id")
	if id == "" {
		id = c.URLParam("id")
	}
	return v.DestroyModelMixin.Destroy(c, id)
}
