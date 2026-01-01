package urls

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterInclude(t *testing.T) {
	// Sub-router (e.g. accounts)
	subRouter := NewRouter()
	subRouter.Get("/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "login page")
	}), "login")

	// Main router
	mainRouter := NewRouter()
	mainRouter.Include("/accounts", subRouter, "accounts")

	// Test request to nested route
	req := httptest.NewRequest("GET", "/accounts/login", nil)
	w := httptest.NewRecorder()

	mainRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "login page" {
		t.Errorf("Expected body 'login page', got '%s'", w.Body.String())
	}
}

func TestRouterReverseNamespaced(t *testing.T) {
	subRouter := NewRouter()
	subRouter.Get("/profile/{id:int}", nil, "profile")

	mainRouter := NewRouter()
	mainRouter.Include("/users", subRouter, "users")

	// Reverse with namespace:name
	url, err := mainRouter.Reverse("users:profile", 123)
	if err != nil {
		t.Fatalf("Reverse failed: %v", err)
	}

	expected := "/users/profile/123"
	if url != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, url)
	}
}

func TestRouterDeepNesting(t *testing.T) {
	level3 := NewRouter()
	level3.Get("/end", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}), "end")

	level2 := NewRouter()
	level2.Include("/l3", level3, "l3")

	level1 := NewRouter()
	level1.Include("/l2", level2, "l2")

	// Path: /l2/l3/end
	req := httptest.NewRequest("GET", "/l2/l3/end", nil)
	w := httptest.NewRecorder()

	level1.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for deep nesting, got %d", w.Code)
	}
}
