package views

import (
	"net/http"
	"reflect"
)

// View defines the interface for an API view
type View interface {
	Dispatch(c *Context) Response
}

// APIViewMethods interface
type APIViewMethods interface {
	Get(c *Context) Response
	Post(c *Context) Response
	Put(c *Context) Response
	Patch(c *Context) Response
	Delete(c *Context) Response
}

// BaseAPIView provides the default dispatch logic
type BaseAPIView struct {
	PermissionClasses []Permission
}

func (v *BaseAPIView) GetPermissionClasses() []Permission {
	return v.PermissionClasses
}

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
	// Check Permissions
	for _, perm := range v.PermissionClasses {
		if !perm.HasPermission(c) {
			return Forbidden("You do not have permission to perform this action.")
		}
	}

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
			User:           r.Context().Value("user"), // Extract from context (string key for simplicity)
		}
		_ = ctx.ParseRequest()

		var resp Response
		// Check if it's a BaseAPIView child that wants default dispatch
		if methods, ok := v.(APIViewMethods); ok {
			// We need to get the PermissionClasses from the actual instance v
			// Since v embeds BaseAPIView, we can use an interface check or reflection
			var permissionClasses []Permission
			if base, ok := v.(interface{ GetPermissionClasses() []Permission }); ok {
				permissionClasses = base.GetPermissionClasses()
			} else if base, ok := v.(*BaseAPIView); ok {
				permissionClasses = base.PermissionClasses
			}

			// If we still don't have them, try reflection as a fallback for embedded structs
			if permissionClasses == nil {
				val := reflect.ValueOf(v)
				if val.Kind() == reflect.Ptr {
					val = val.Elem()
				}
				if val.Kind() == reflect.Struct {
					field := val.FieldByName("BaseAPIView")
					if field.IsValid() && field.Kind() == reflect.Struct {
						permField := field.FieldByName("PermissionClasses")
						if permField.IsValid() && permField.Kind() == reflect.Slice {
							permissionClasses = permField.Interface().([]Permission)
						}
					}
				}
			}

			disp := &BaseAPIView{PermissionClasses: permissionClasses}
			resp = disp.Dispatch(ctx, methods)
		} else if dv, ok := v.(interface {
			Dispatch(c *Context) Response
		}); ok {
			resp = dv.Dispatch(ctx)
		} else {
			resp = MethodNotAllowed()
		}

		resp.Render(w)
	})
}
