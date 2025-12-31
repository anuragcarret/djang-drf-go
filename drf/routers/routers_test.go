package routers

import (
	"testing"

	"github.com/anuragcarret/djang-drf-go/drf/views"
)

type MockViewSet struct {
	views.BaseAPIView
}

func (v *MockViewSet) List(c *views.Context) views.Response { return views.OK(nil) }

func TestDefaultRouter(t *testing.T) {
	t.Run("registers viewset and generates urls", func(t *testing.T) {
		router := NewDefaultRouter()
		router.Register("users", &MockViewSet{})

		patterns := router.URLs()
		if len(patterns) == 0 {
			t.Fatal("expected generated url patterns, got 0")
		}

		// Check for list endpoint
		foundList := false
		for _, p := range patterns {
			if p.Pattern == "users/" {
				foundList = true
				break
			}
		}
		if !foundList {
			t.Error("missing 'users/' pattern")
		}
	})

	t.Run("generates detail urls with regex", func(t *testing.T) {
		router := NewDefaultRouter()
		router.Register("users", &MockViewSet{})

		patterns := router.URLs()
		foundDetail := false
		for _, p := range patterns {
			// Detail pattern typically has a parameter like {id} or {pk}
			// In core/urls it might be users/<int:id>/
			if p.Pattern == "users/<int:id>/" {
				foundDetail = true
				break
			}
		}
		if !foundDetail {
			t.Error("missing 'users/<int:id>/' pattern")
		}
	})
}
