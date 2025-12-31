package urls

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter(t *testing.T) {
	t.Run("matches simple path", func(t *testing.T) {
		r := NewRouter()
		handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("home"))
		})
		handler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("about"))
		})

		r.Path("/", handler1, "home")
		r.Path("/about", handler2, "about")

		tests := []struct {
			path     string
			expected string
		}{
			{"/", "home"},
			{"/about", "about"},
		}

		for _, tc := range tests {
			req := httptest.NewRequest("GET", tc.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Body.String() != tc.expected {
				t.Errorf("path %s: expected %s, got %s", tc.path, tc.expected, rec.Body.String())
			}
		}
	})

	t.Run("handles 404", func(t *testing.T) {
		r := NewRouter()
		req := httptest.NewRequest("GET", "/unknown", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("matches path parameters", func(t *testing.T) {
		r := NewRouter()
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, _ := URLParam[int](r, "id")
			w.Write([]byte("user " + string(rune('0'+id)))) // primitive conversion for testing
		})

		r.Path("/users/{id:int}", handler, "user-detail")

		req := httptest.NewRequest("GET", "/users/1", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Body.String() != "user 1" {
			t.Errorf("expected 'user 1', got '%s'", rec.Body.String())
		}
	})

	t.Run("handles method-specific routes", func(t *testing.T) {
		r := NewRouter()
		r.Get("/items", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("list"))
		}), "list")
		r.Post("/items", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("create"))
		}), "create")

		// Test GET
		req := httptest.NewRequest("GET", "/items", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Body.String() != "list" {
			t.Errorf("GET: expected list, got %s", rec.Body.String())
		}

		// Test POST
		req = httptest.NewRequest("POST", "/items", nil)
		rec = httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Body.String() != "create" {
			t.Errorf("POST: expected create, got %s", rec.Body.String())
		}

		// Test 405
		req = httptest.NewRequest("PUT", "/items", nil)
		rec = httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", rec.Code)
		}
	})

	t.Run("handles nested routers with Include", func(t *testing.T) {
		main := NewRouter()
		app := NewRouter()
		app.Get("/info", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("app-info"))
		}), "info")

		main.Include("/app", app, "app")

		req := httptest.NewRequest("GET", "/app/info", nil)
		rec := httptest.NewRecorder()
		main.ServeHTTP(rec, req)

		if rec.Body.String() != "app-info" {
			t.Errorf("expected 'app-info', got '%s'", rec.Body.String())
		}
	})
}

func TestReverse(t *testing.T) {
	r := NewRouter()
	r.Path("/users/{id:int}", nil, "user-detail")
	r.Path("/articles/{slug:slug}", nil, "article-detail")

	app := NewRouter()
	app.Get("/{id:int}", nil, "item")
	r.Include("/items", app, "api")

	tests := []struct {
		name     string
		args     []interface{}
		expected string
	}{
		{"user-detail", []interface{}{42}, "/users/42"},
		{"article-detail", []interface{}{"hello-world"}, "/articles/hello-world"},
		{"api:item", []interface{}{10}, "/items/10"},
	}

	for _, tc := range tests {
		url, err := r.Reverse(tc.name, tc.args...)
		if err != nil {
			t.Errorf("Reverse(%s) failed: %v", tc.name, err)
			continue
		}
		if url != tc.expected {
			t.Errorf("Reverse(%s): expected %s, got %s", tc.name, tc.expected, url)
		}
	}
}
