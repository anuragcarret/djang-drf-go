package admin

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/anuragcarret/djang-drf-go/contrib/auth"
	"github.com/anuragcarret/djang-drf-go/orm/db"
	"github.com/anuragcarret/djang-drf-go/orm/queryset"
)

// TestAddViewRoute tests that the add view route is accessible
func TestAddViewRoute(t *testing.T) {
	t.Skip("Add view not yet implemented - TDD red phase")

	database := setupTestDB(t)
	defer database.Close()

	// Register a test model
	Register[*auth.User](DefaultSite, &ModelAdmin{
		ListDisplay: []string{"ID", "Username", "Email"},
	})

	// Get admin routes
	router := DefaultSite.URLs(database)

	// Test GET /admin/auth/User/add/
	req := httptest.NewRequest("GET", "/auth/User/add/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()

	// Verify form is rendered
	if !strings.Contains(body, "<form") {
		t.Error("Expected form element in response")
	}

	// Verify form action points to correct URL
	if !strings.Contains(body, `action="/admin/auth/User/add/"`) {
		t.Error("Expected form action attribute")
	}

	// Verify username field is present
	if !strings.Contains(body, "username") {
		t.Error("Expected username field in form")
	}

	// Verify email field is present
	if !strings.Contains(body, "email") {
		t.Error("Expected email field in form")
	}
}

// TestAddViewSubmit tests form submission creates a record
func TestAddViewSubmit(t *testing.T) {
	t.Skip("Add view not yet implemented - TDD red phase")

	database := setupTestDB(t)
	defer database.Close()

	Register[*auth.User](DefaultSite, &ModelAdmin{})

	router := DefaultSite.URLs(database)

	// Prepare form data
	formData := url.Values{}
	formData.Set("username", "testuser")
	formData.Set("email", "test@example.com")
	formData.Set("password", "testpass123")

	req := httptest.NewRequest("POST", "/auth/User/add/", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should redirect after successful creation
	if w.Code != http.StatusFound && w.Code != http.StatusSeeOther {
		t.Errorf("Expected redirect (302/303), got %d", w.Code)
	}

	// Verify redirect location
	location := w.Header().Get("Location")
	if !strings.Contains(location, "/change/") {
		t.Errorf("Expected redirect to change form, got: %s", location)
	}

	// Verify record was created in database
	qs := queryset.NewQuerySet[*auth.User](database)
	users, err := qs.Filter(queryset.Q{"username": "testuser"}).All()
	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user, found %d", len(users))
	}

	if users[0].Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", users[0].Username)
	}
}

// TestAddViewValidation tests that validation errors are displayed
func TestAddViewValidation(t *testing.T) {
	t.Skip("Add view not yet implemented - TDD red phase")

	database := setupTestDB(t)
	defer database.Close()

	Register[*auth.User](DefaultSite, &ModelAdmin{})

	router := DefaultSite.URLs(database)

	// Submit with missing required field
	formData := url.Values{}
	formData.Set("email", "test@example.com")
	// Missing username (required field)

	req := httptest.NewRequest("POST", "/auth/User/add/", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 200 with error message (not redirect)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for validation error, got %d", w.Code)
	}

	body := w.Body.String()

	// Verify error message is displayed
	if !strings.Contains(body, "error") && !strings.Contains(body, "required") {
		t.Error("Expected validation error message in response")
	}
}

// TestFieldTypeDetection tests that different field types render correctly
func TestFieldTypeDetection(t *testing.T) {
	t.Skip("Add view not yet implemented - TDD red phase")

	database := setupTestDB(t)
	defer database.Close()

	Register[*auth.User](DefaultSite, &ModelAdmin{})

	router := DefaultSite.URLs(database)
	req := httptest.NewRequest("GET", "/auth/User/add/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	body := w.Body.String()

	// Test string field -> text input
	if !strings.Contains(body, `type="text"`) {
		t.Error("Expected text input for string fields")
	}

	// Test boolean field -> checkbox
	if !strings.Contains(body, `type="checkbox"`) {
		t.Error("Expected checkbox for boolean fields")
	}
}

// Helper function to set up test database
func setupTestDB(t *testing.T) *db.DB {
	// Use PostgreSQL test database instead of SQLite
	database, err := db.NewDB("postgres", "user=demo_user password=postgres dbname=demo_db sslmode=disable")
	if err != nil {
		t.Skipf("Skipping test - cannot connect to test database: %v", err)
	}

	// Clean up test data
	_, _ = database.Exec("DELETE FROM go_users WHERE username LIKE 'test%'")

	return database
}
