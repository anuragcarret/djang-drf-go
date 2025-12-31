package views

import (
	"net/http"
)

// View defines the interface for an API view
type View interface {
	Dispatch(c *Context) Response
}

// APIView methods interface
type APIViewMethods interface {
	Get(c *Context) Response
	Post(c *Context) Response
	Put(c *Context) Response
	Patch(c *Context) Response
	Delete(c *Context) Response
}

// BaseAPIView provides the default dispatch logic
type BaseAPIView struct{}

func (v *BaseAPIView) Response(w http.ResponseWriter) *ResponseWrapper {
	return &ResponseWrapper{w: w}
}

type ResponseWrapper struct {
	w http.ResponseWriter
}

func (rw *ResponseWrapper) OK(data interface{})      { OK(data).Render(rw.w) }
func (rw *ResponseWrapper) Created(data interface{}) { Created(data).Render(rw.w) }
func (rw *ResponseWrapper) BadRequest(msg string) {
	BadRequest(map[string]string{"error": msg}).Render(rw.w)
}
func (rw *ResponseWrapper) NotFound(msg string) { NotFound(msg).Render(rw.w) }

func (v *BaseAPIView) Get(c *Context) Response    { return MethodNotAllowed() }
func (v *BaseAPIView) Post(c *Context) Response   { return MethodNotAllowed() }
func (v *BaseAPIView) Put(c *Context) Response    { return MethodNotAllowed() }
func (v *BaseAPIView) Patch(c *Context) Response  { return MethodNotAllowed() }
func (v *BaseAPIView) Delete(c *Context) Response { return MethodNotAllowed() }

func (v *BaseAPIView) Dispatch(c *Context, methods APIViewMethods) Response {
	switch c.Request.Method {
	case "GET":
		return methods.Get(c)
	case "POST":
		return methods.Post(c)
	case "PUT":
		return methods.Put(c)
	case "PATCH":
		return methods.Patch(c)
	case "DELETE":
		return methods.Delete(c)
	default:
		return MethodNotAllowed()
	}
}

// Handler converts a View into an http.Handler
func Handler(v interface{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := &Context{
			Request:        r,
			ResponseWriter: w,
			Query:          r.URL.Query(),
			Data:           make(map[string]interface{}),
		}
		_ = ctx.ParseRequest()

		var resp Response
		if dv, ok := v.(interface {
			Dispatch(c *Context) Response
		}); ok {
			// If View implements Dispatch directly (user-defined)
			resp = dv.Dispatch(ctx)
		} else if methods, ok := v.(APIViewMethods); ok {
			// If it's a BaseAPIView child
			base := BaseAPIView{}
			resp = base.Dispatch(ctx, methods)
		} else {
			resp = MethodNotAllowed()
		}

		resp.Render(w)
	})
}
