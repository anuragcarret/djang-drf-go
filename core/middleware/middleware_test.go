package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChain(t *testing.T) {
	t.Run("executes middleware in order", func(t *testing.T) {
		var order []string

		m1 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "m1-start")
				next.ServeHTTP(w, r)
				order = append(order, "m1-end")
			})
		}

		m2 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "m2-start")
				next.ServeHTTP(w, r)
				order = append(order, "m2-end")
			})
		}

		final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "final")
			w.WriteHeader(http.StatusOK)
		})

		chain := New(m1, m2)
		handler := chain.Then(final)

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		expected := []string{"m1-start", "m2-start", "final", "m2-end", "m1-end"}
		if len(order) != len(expected) {
			t.Fatalf("expected %d calls, got %d", len(expected), len(order))
		}

		for i, v := range expected {
			if order[i] != v {
				t.Errorf("at index %d: expected %s, got %s", i, v, order[i])
			}
		}
	})

	t.Run("ThenFunc works correctly", func(t *testing.T) {
		m := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Middleware", "true")
				next.ServeHTTP(w, r)
			})
		}

		chain := New(m)
		handler := chain.ThenFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
		})

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Errorf("expected 202, got %d", rec.Code)
		}
		if rec.Header().Get("X-Middleware") != "true" {
			t.Error("middleware header not set")
		}
	})

	t.Run("Append adds to existing chain", func(t *testing.T) {
		var order []string
		m1 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "m1")
				next.ServeHTTP(w, r)
			})
		}
		m2 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "m2")
				next.ServeHTTP(w, r)
			})
		}

		chain := New(m1).Append(m2)
		handler := chain.ThenFunc(func(w http.ResponseWriter, r *http.Request) {})

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if order[0] != "m1" || order[1] != "m2" {
			t.Errorf("expected [m1, m2], got %v", order)
		}
	})
}
