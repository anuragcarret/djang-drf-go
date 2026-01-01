package views

import (
	"encoding/json"
	"net/http"

	"github.com/anuragcarret/djang-drf-go/drf/serializers"
)

// Response represents an HTTP response
type Response struct {
	Status      int
	Data        interface{}
	Headers     map[string]string
	ContentType string
	Depth       int
}

func (r Response) Render(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json") // Default
	if r.ContentType != "" {
		w.Header().Set("Content-Type", r.ContentType)
	}
	for k, v := range r.Headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(r.Status)
	if r.Data != nil {
		// Automatically serialize data to strip write_only fields
		serialized := serializers.Serialize(r.Data, r.Depth)
		json.NewEncoder(w).Encode(serialized)
	}
}

// Response helpers
func OK(data interface{}) Response {
	return Response{Status: http.StatusOK, Data: data}
}

func Created(data interface{}) Response {
	return Response{Status: http.StatusCreated, Data: data}
}

func NoContent() Response {
	return Response{Status: http.StatusNoContent}
}

func BadRequest(data interface{}) Response {
	return Response{Status: http.StatusBadRequest, Data: data}
}

func NotFound(msg string) Response {
	return Response{Status: http.StatusNotFound, Data: map[string]string{"detail": msg}}
}

func Forbidden(msg string) Response {
	return Response{Status: http.StatusForbidden, Data: map[string]string{"detail": msg}}
}

func MethodNotAllowed() Response {
	return Response{Status: http.StatusMethodNotAllowed, Data: map[string]string{"detail": "Method not allowed"}}
}
