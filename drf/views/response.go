package views

import (
	"encoding/json"
	"net/http"
)

// Response represents an HTTP response
type Response struct {
	Status      int
	Data        interface{}
	Headers     map[string]string
	ContentType string
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
		json.NewEncoder(w).Encode(r.Data)
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

func MethodNotAllowed() Response {
	return Response{Status: http.StatusMethodNotAllowed, Data: map[string]string{"detail": "Method not allowed"}}
}
