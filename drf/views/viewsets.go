package views

import (
	"net/http"

	"github.com/anuragcarret/djang-drf-go/drf/pagination"
	"github.com/anuragcarret/djang-drf-go/orm/db"
	"github.com/anuragcarret/djang-drf-go/orm/queryset"
)

// ModelViewSet provides default CRUD actions for a model
type ModelViewSet[T queryset.ModelInterface] struct {
	BaseAPIView
	DB         *db.DB
	Pagination *pagination.PageNumberPagination
	Depth      int
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
	// Logic to save T from c.Data
	// Generic implementation would use reflection to populate struct
	return Created(map[string]string{"status": "Created (Generic)"})
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

// ModelHandler converts a ModelViewSet into an http.Handler with CRUD routing
func ModelHandler[T queryset.ModelInterface](v *ModelViewSet[T]) http.Handler {
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
		default:
			resp = MethodNotAllowed()
		}

		// Apply ViewSet depth to response if not already set
		if resp.Depth == 0 && v.Depth > 0 {
			resp.Depth = v.Depth
		}

		resp.Render(w)
	})
}
