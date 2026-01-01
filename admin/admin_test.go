package admin

import (
	"testing"
)

// MockModel is defined in admin_integration_test.go

func TestAdminSite_Registration(t *testing.T) {
	site := NewAdminSite()
	adminConfig := &ModelAdmin{}

	t.Run("registers model with custom admin", func(t *testing.T) {
		err := Register[*MockModel](site, adminConfig)
		if err != nil {
			t.Fatalf("Failed to register model: %v", err)
		}

		// Since GetAdmin is removed/internal, we assume success if no error
		// We could inspect internal registry if needed via reflection or just rely on error
	})

	t.Run("fails on duplicate registration", func(t *testing.T) {
		err := Register[*MockModel](site, adminConfig)
		if err == nil {
			t.Error("Expected error for duplicate registration, got nil")
		}
	})
}
