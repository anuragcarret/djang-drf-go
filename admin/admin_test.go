package admin

import (
	"testing"
)

type MockModel struct {
	ID   int
	Name string
}

type MockAdmin struct {
	ModelAdmin
}

func TestAdminSite_Registration(t *testing.T) {
	site := NewAdminSite()
	model := &MockModel{}
	admin := &MockAdmin{}

	t.Run("registers model with custom admin", func(t *testing.T) {
		err := site.Register(model, admin)
		if err != nil {
			t.Fatalf("Failed to register model: %v", err)
		}

		retrieved, err := site.GetAdmin(model)
		if err != nil {
			t.Fatalf("Failed to retrieve admin: %v", err)
		}

		if retrieved != admin {
			t.Error("Retrieved admin does not match registered admin")
		}
	})

	t.Run("fails on duplicate registration", func(t *testing.T) {
		err := site.Register(model, admin)
		if err == nil {
			t.Error("Expected error for duplicate registration, got nil")
		}
	})
}
