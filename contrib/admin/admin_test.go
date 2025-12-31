package admin

import (
	"testing"
)

type MockModel struct{}

func (m *MockModel) TableName() string { return "mock_table" }
func (m *MockModel) Meta() interface{} { return nil }

type MockAdmin struct {
	ModelAdmin
}

func TestAdminRegistration(t *testing.T) {
	t.Run("registers model with custom admin", func(t *testing.T) {
		site := NewAdminSite("admin", "/admin")
		model := &MockModel{}
		adm := &MockAdmin{}

		site.Register(model, adm)

		if _, ok := site.Models["mock_table"]; !ok {
			t.Error("model not registered in site")
		}
	})

	t.Run("generates admin urls", func(t *testing.T) {
		site := NewAdminSite("admin", "/admin")
		site.Register(&MockModel{}, nil)

		patterns := site.URLs()
		if len(patterns) == 0 {
			t.Fatal("expected admin url patterns, got 0")
		}

		foundList := false
		for _, p := range patterns {
			if p.Pattern == "mock_table/" {
				foundList = true
				break
			}
		}
		if !foundList {
			t.Error("missing 'mock_table/' admin pattern")
		}
	})
}
