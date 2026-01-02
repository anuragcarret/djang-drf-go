package urls

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Tests for path parameter support in router

func TestRouterPathParams_Basic(t *testing.T) {
	r := NewRouter()

	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		called = true

		// Extract parameter
		params := GetParams(req)
		if params == nil {
			t.Error("Expected params to be available")
			return
		}

		id := params.Get("id")
		if id != "123" {
			t.Errorf("Expected id=123, got %s", id)
		}
	})

	r.Get("/users/{id}/", handler, "user-detail")

	// Should match
	req := httptest.NewRequest("GET", "/users/123/", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if !called {
		t.Error("Handler was not called")
	}
}

func TestRouterPathParams_MultipleParams(t *testing.T) {
	r := NewRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		params := GetParams(req)

		if params.Get("app") != "posts" {
			t.Errorf("Expected app=posts, got %s", params.Get("app"))
		}

		if params.Get("model") != "Post" {
			t.Errorf("Expected model=Post, got %s", params.Get("model"))
		}

		if params.Get("id") != "456" {
			t.Errorf("Expected id=456, got %s", params.Get("id"))
		}
	})

	r.Get("/{app}/{model}/{id}/change/", handler, "change")

	req := httptest.NewRequest("GET", "/posts/Post/456/change/", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
}

func TestRouterPathParams_TypeConstraintInt(t *testing.T) {
	r := NewRouter()

	matchedInt := false
	r.Get("/users/{id:int}/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		matchedInt = true
		params := GetParams(req)

		// Should be able to get as int
		id, err := params.GetInt("id")
		if err != nil {
			t.Errorf("Failed to get id as int: %v", err)
		}
		if id != 123 {
			t.Errorf("Expected id=123, got %d", id)
		}
	}), "user-int")

	// Should match numeric ID
	req1 := httptest.NewRequest("GET", "/users/123/", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	if !matchedInt {
		t.Error("Should have matched /users/123/")
	}

	// Should NOT match non-numeric ID
	matchedInt = false
	req2 := httptest.NewRequest("GET", "/users/abc/", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if matchedInt {
		t.Error("Should not have matched /users/abc/ with int constraint")
	}

	// Should return 404
	if w2.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for non-int param, got %d", w2.Code)
	}
}

func TestRouterPathParams_TypeConstraintString(t *testing.T) {
	r := NewRouter()

	matched := false
	r.Get("/posts/{slug:string}/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		matched = true
		params := GetParams(req)

		if params.Get("slug") != "my-awesome-post" {
			t.Errorf("Expected slug=my-awesome-post, got %s", params.Get("slug"))
		}
	}), "post-slug")

	req := httptest.NewRequest("GET", "/posts/my-awesome-post/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if !matched {
		t.Error("Should have matched /posts/my-awesome-post/")
	}
}

func TestRouterPathParams_Wildcard(t *testing.T) {
	r := NewRouter()

	matched := false
	r.Get("/static/*filepath", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		matched = true
		params := GetParams(req)

		filepath := params.Get("filepath")
		if filepath != "css/style.css" {
			t.Errorf("Expected filepath=css/style.css, got %s", filepath)
		}
	}), "static")

	req := httptest.NewRequest("GET", "/static/css/style.css", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if !matched {
		t.Error("Should have matched wildcard pattern")
	}
}

func TestRouterPathParams_NoMatch(t *testing.T) {
	r := NewRouter()

	r.Get("/users/{id:int}/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		t.Error("Handler should not be called for non-matching route")
	}), "user")

	// Different path entirely
	req := httptest.NewRequest("GET", "/posts/123/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", w.Code)
	}
}

func TestRouterPathParams_PriorityExactMatch(t *testing.T) {
	r := NewRouter()

	exactCalled := false
	paramCalled := false

	// Register exact match first
	r.Get("/users/me/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		exactCalled = true
	}), "user-me")

	// Register param match
	r.Get("/users/{id}/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		paramCalled = true
	}), "user-id")

	// Should match exact route, not param route
	req := httptest.NewRequest("GET", "/users/me/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if !exactCalled {
		t.Error("Exact match should have been called")
	}
	if paramCalled {
		t.Error("Param match should not have been called")
	}
}

func TestRouterPathParams_GetIntUint(t *testing.T) {
	r := NewRouter()

	r.Get("/test/{id}/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		params := GetParams(req)

		// Test GetInt
		id, err := params.GetInt("id")
		if err != nil {
			t.Errorf("GetInt failed: %v", err)
		}
		if id != 42 {
			t.Errorf("Expected id=42, got %d", id)
		}

		// Test GetUint
		uid, err := params.GetUint("id")
		if err != nil {
			t.Errorf("GetUint failed: %v", err)
		}
		if uid != 42 {
			t.Errorf("Expected uid=42, got %d", uid)
		}
	}), "test")

	req := httptest.NewRequest("GET", "/test/42/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
}

func TestRouterPathParams_EmptyParams(t *testing.T) {
	r := NewRouter()

	r.Get("/simple/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		params := GetParams(req)

		// Should still be accessible but empty
		if params == nil {
			t.Error("Params should not be nil even for routes without params")
		}

		if params.Get("nonexistent") != "" {
			t.Error("Expected empty string for non-existent param")
		}
	}), "simple")

	req := httptest.NewRequest("GET", "/simple/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
}

func TestRouterPathParams_TrailingSlash(t *testing.T) {
	r := NewRouter()

	matched := false
	r.Get("/users/{id}/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		matched = true
	}), "user")

	// With trailing slash
	req1 := httptest.NewRequest("GET", "/users/123/", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	if !matched {
		t.Error("Should match with trailing slash")
	}

	// Without trailing slash (should also match or redirect)
	matched = false
	req2 := httptest.NewRequest("GET", "/users/123", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	// Either should match or redirect to trailing slash version
	if !matched && w2.Code != http.StatusMovedPermanently && w2.Code != http.StatusNotFound {
		t.Errorf("Expected match, redirect, or 404; got %d", w2.Code)
	}
}
