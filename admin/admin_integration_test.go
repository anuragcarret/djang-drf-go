package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anuragcarret/djang-drf-go/core/apps"
	"github.com/anuragcarret/djang-drf-go/orm/db"
)

// MockDB for testing
var mockDB *db.DB

func TestAdminRouter(t *testing.T) {
	// Setup App Registry for testing
	mockApp := &MockApp{}
	// Ignore error if already registered (in case checking conflicts)
	_ = apps.Apps.Register(mockApp)
	// We might need to manually inject the model if Populate isn't called or to ensure it's there
	// But RegisterModel is usually called by Populate or manually
	// MockApp.AppConfig() should return the model mapping

	// Force populate to ensure registry is ready (idempotent-ish)
	_ = apps.Apps.Populate()

	// Manually register the model to the app in registry to be sure (since MockApp might be simple)
	_ = apps.Apps.RegisterModel("mock_app", &MockModel{}, "mock_models")

	site := NewAdminSite()
	adminConfig := &ModelAdmin{
		ListDisplay: []string{"Name"},
	}

	err := Register[*MockModel](site, adminConfig)
	if err != nil {
		t.Fatalf("Failed to register model: %v", err)
	}

	router := site.URLs(mockDB) // Inject DB dependency

	// Test Metadata/Index Endpoint
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for admin index, got %d", w.Code)
		t.Logf("Response: %s", w.Body.String())
	}

	// Test Model List Endpoint (URL structure dependent on implementation)
	// Assuming /app_label/model_name pattern
	// Since MockModel isn't in a real app, we might need to mock AppConfig lookups or use a simplified structure
	// Test List Endpoint for MockModel
	// Path should be /mock_app/MockModel/
	reqList := httptest.NewRequest("GET", "/mock_app/MockModel", nil)
	wList := httptest.NewRecorder()
	router.ServeHTTP(wList, reqList)

	// Since mockDB is nil, we expect ServiceUnavailable
	if wList.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected 503 ServiceUnavailable for model list (nil db), got %d", wList.Code)
		t.Logf("Response: %s", wList.Body.String())
	}
}

func TestAdminURLs_Structure(t *testing.T) {
	site := NewAdminSite()
	Register[*MockModel](site, &ModelAdmin{})

	// Check if router contains expected patterns
	// This requires inspecting the router internals or integration testing with known paths
}

type MockApp struct{}

func (a *MockApp) AppConfig() *apps.AppConfig {
	return &apps.AppConfig{
		Name:  "mock_app",
		Label: "mock_app",
		Models: map[string]interface{}{
			"mock_models": &MockModel{},
		},
	}
}

func (a *MockApp) Ready() error { return nil }

type MockModel struct {
	ID   int
	Name string
}

func (m *MockModel) TableName() string { return "mock_models" }
