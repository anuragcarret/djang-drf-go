package views

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type TestView struct {
	BaseAPIView
}

func (v *TestView) Get(c *Context) Response {
	return OK(map[string]string{"message": "hello"})
}

func (v *TestView) Post(c *Context) Response {
	return Created(c.Data)
}

func TestAPIView(t *testing.T) {
	t.Run("handles GET request", func(t *testing.T) {
		view := &TestView{}
		handler := Handler(view)

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		expectedBody := `{"message":"hello"}`
		if strings.TrimSpace(rec.Body.String()) != expectedBody {
			t.Errorf("expected %s, got %s", expectedBody, rec.Body.String())
		}
	})

	t.Run("handles POST request with data", func(t *testing.T) {
		view := &TestView{}
		handler := Handler(view)

		body := `{"name":"test"}`
		req := httptest.NewRequest("POST", "/", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"name":"test"`) {
			t.Errorf("missing data in response: %s", rec.Body.String())
		}
	})

	t.Run("returns 405 for unimplemented method", func(t *testing.T) {
		view := &TestView{}
		handler := Handler(view)

		req := httptest.NewRequest("PUT", "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", rec.Code)
		}
	})
}
