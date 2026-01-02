package views

import (
	"net/http"
	"reflect"

	"github.com/anuragcarret/djang-drf-go/drf/pagination"
	"github.com/anuragcarret/djang-drf-go/orm/db"
	"github.com/anuragcarret/djang-drf-go/orm/queryset"
)

// APIViewSet defines the interface for a ViewSet to allow overriding in substructs
type APIViewSet interface {
	List(c *Context) Response
	Create(c *Context) Response
	Retrieve(c *Context, id string) Response
	Update(c *Context, id string) Response
	Delete(c *Context, id string) Response
	GetDepth() int
}

// ModelViewSet provides default CRUD actions for a model
type ModelViewSet[T queryset.ModelInterface] struct {
	BaseAPIView
	DB         *db.DB
	Pagination *pagination.PageNumberPagination
	Depth      int
}

func (v *ModelViewSet[T]) GetDepth() int {
	return v.Depth
}

func (v *ModelViewSet[T]) List(c *Context) Response {
	qs := queryset.NewQuerySet[T](v.DB)
	qs = ApplyFilters(qs, c.Query)

	// Basic total count for pagination
	// (In real ORM, this would be a Count() method on QuerySet)
	count := 10 // Placeholder

	results, err := v.paginate(qs, c)
	if err != nil {
		return BadRequest(map[string]string{"error": err.Error()})
	}

	if v.Pagination != nil {
		page := v.Pagination.GetPaginatedResponse(results, count, c.Request)
		return OK(page)
	}

	return OK(results)
}

func (v *ModelViewSet[T]) Create(c *Context) Response {
	var instance T
	// For pointer types, we need to initialize
	tType := reflect.TypeOf(instance)
	if tType.Kind() == reflect.Ptr {
		instance = reflect.New(tType.Elem()).Interface().(T)
	}

	if err := c.Bind(instance); err != nil {
		return BadRequest(map[string]string{"error": "Failed to bind data: " + err.Error()})
	}

	if err := v.PerformCreate(c, instance); err != nil {
		return BadRequest(map[string]string{"error": "Failed to create: " + err.Error()})
	}

	return Created(instance)
}

func (v *ModelViewSet[T]) PerformCreate(c *Context, obj T) error {
	qs := queryset.NewQuerySet[T](v.DB)
	return qs.Create(obj)
}

func (v *ModelViewSet[T]) Retrieve(c *Context, id string) Response {
	return OK(map[string]string{"id": id, "status": "Retrieved (Generic)"})
}

func (v *ModelViewSet[T]) Update(c *Context, id string) Response {
	return OK(map[string]string{"id": id, "status": "Updated (Generic)"})
}

func (v *ModelViewSet[T]) Delete(c *Context, id string) Response {
	return NoContent()
}

// Helpers
func (v *ModelViewSet[T]) paginate(qs *queryset.QuerySet[T], c *Context) ([]T, error) {
	if v.Pagination == nil {
		return qs.All()
	}
	// TODO: Apply Limit/Offset based on page
	return qs.All()
}

// ModelHandler converts an APIViewSet into an http.Handler with CRUD routing
func ModelHandler(v APIViewSet) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := &Context{
			Request:        r,
			ResponseWriter: w,
			Query:          r.URL.Query(),
			Data:           make(map[string]interface{}),
			User:           r.Context().Value("user"),
		}
		_ = ctx.ParseRequest()

		var resp Response
		switch r.Method {
		case "GET":
			id := r.URL.Query().Get("id")
			if id != "" {
				resp = v.Retrieve(ctx, id)
			} else {
				resp = v.List(ctx)
			}
		case "POST":
			resp = v.Create(ctx)
		case "PUT", "PATCH":
			id := r.URL.Query().Get("id")
			resp = v.Update(ctx, id)
		case "DELETE":
			id := r.URL.Query().Get("id")
			resp = v.Delete(ctx, id)
		default:
			resp = MethodNotAllowed()
		}

		// Apply ViewSet depth to response if not already set
		if resp.Depth == 0 && v.GetDepth() > 0 {
			resp.Depth = v.GetDepth()
		}

		resp.Render(w)
	})
}
