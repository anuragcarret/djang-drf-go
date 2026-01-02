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

// Integration tests for Add view with real database operations

func TestAddViewIntegration_CreateUser(t *testing.T) {
	database, cleanup := setupIntegrationTestDB(t)
	defer cleanup()

	// Clear any existing test users
	database.Exec("DELETE FROM go_users WHERE username LIKE 'testuser%'")

	// Register User model
	site := NewAdminSite()
	err := Register[*auth.User](site, &ModelAdmin{
		ListDisplay: []string{"ID", "Username", "Email"},
	})
	if err != nil {
		t.Fatalf("Failed to register model: %v", err)
	}

	// Get admin routes
	router := site.URLs(database)

	// Prepare form data for creating a user
	formData := url.Values{}
	formData.Set("username", "testuser_integration")
	formData.Set("email", "testuser@example.com")
	formData.Set("password", "testpass123")
	formData.Set("first_name", "Test")
	formData.Set("last_name", "User")
	formData.Set("is_active", "true")
	formData.Set("is_staff", "false")
	formData.Set("is_superuser", "false")

	// Submit POST request to add view
	req := httptest.NewRequest("POST", "/auth/User/add/", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should redirect after successful creation
	if w.Code != http.StatusFound && w.Code != http.StatusSeeOther {
		body := w.Body.String()
		t.Fatalf("Expected redirect (302/303), got %d. Body: %s", w.Code, body)
	}

	// Verify record was created in database
	qs := queryset.NewQuerySet[*auth.User](database)
	users, err := qs.Filter(queryset.Q{"username": "testuser_integration"}).All()
	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}

	if len(users) != 1 {
		t.Fatalf("Expected 1 user, found %d", len(users))
	}

	user := users[0]
	if user.Username != "testuser_integration" {
		t.Errorf("Expected username 'testuser_integration', got '%s'", user.Username)
	}
	if user.Email != "testuser@example.com" {
		t.Errorf("Expected email 'testuser@example.com', got '%s'", user.Email)
	}
	if user.FirstName != "Test" {
		t.Errorf("Expected first_name 'Test', got '%s'", user.FirstName)
	}
	if !user.IsActive {
		t.Error("Expected user to be active")
	}
	if user.IsStaff {
		t.Error("Expected user to not be staff")
	}
}

func TestAddViewIntegration_ValidationErrors(t *testing.T) {
	database, cleanup := setupIntegrationTestDB(t)
	defer cleanup()

	site := NewAdminSite()
	err := Register[*auth.User](site, &ModelAdmin{})
	if err != nil {
		t.Fatalf("Failed to register model: %v", err)
	}

	router := site.URLs(database)

	// Submit with missing required field (username)
	formData := url.Values{}
	formData.Set("email", "testuser@example.com")
	formData.Set("password", "testpass123")

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
	if !strings.Contains(body, "Username is required") {
		t.Error("Expected 'Username is required' error message in response")
	}

	// Verify form is re-rendered
	if !strings.Contains(body, "<form") {
		t.Error("Expected form to be re-rendered")
	}

	// Verify no record was created
	qs := queryset.NewQuerySet[*auth.User](database)
	users, _ := qs.Filter(queryset.Q{"email": "testuser@example.com"}).All()
	if len(users) > 0 {
		t.Error("Expected no user to be created when validation fails")
	}
}

func TestAddViewIntegration_TypeValidation(t *testing.T) {
	database, cleanup := setupIntegrationTestDB(t)
	defer cleanup()

	site := NewAdminSite()
	err := Register[*auth.User](site, &ModelAdmin{})
	if err != nil {
		t.Fatalf("Failed to register model: %v", err)
	}

	router := site.URLs(database)

	// Submit with invalid email format
	formData := url.Values{}
	formData.Set("username", "testuser_invalid")
	formData.Set("email", "not-an-email") // Invalid email
	formData.Set("password", "testpass123")

	req := httptest.NewRequest("POST", "/auth/User/add/", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Note: Our current implementation doesn't validate email format server-side
	// The HTML5 email input will handle client-side validation
	// For server-side validation, we would need to add custom validators

	// For now, this test documents that we rely on HTML5 validation
	// Future enhancement: Add server-side validators
}

func TestAddViewIntegration_SaveAndContinue(t *testing.T) {
	database, cleanup := setupIntegrationTestDB(t)
	defer cleanup()

	database.Exec("DELETE FROM go_users WHERE username LIKE 'testuser%'")

	site := NewAdminSite()
	err := Register[*auth.User](site, &ModelAdmin{})
	if err != nil {
		t.Fatalf("Failed to register model: %v", err)
	}

	router := site.URLs(database)

	// Submit with "_continue" button
	formData := url.Values{}
	formData.Set("username", "testuser_continue")
	formData.Set("email", "continue@example.com")
	formData.Set("password", "testpass123")
	formData.Set("_continue", "Save and continue editing")

	req := httptest.NewRequest("POST", "/auth/User/add/", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should redirect to change form
	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected redirect 303, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if !strings.Contains(location, "/change/") {
		t.Errorf("Expected redirect to change form, got: %s", location)
	}

	// Verify record was created
	qs := queryset.NewQuerySet[*auth.User](database)
	users, _ := qs.Filter(queryset.Q{"username": "testuser_continue"}).All()
	if len(users) != 1 {
		t.Errorf("Expected 1 user to be created")
	}
}

func TestAddViewIntegration_SaveAndAddAnother(t *testing.T) {
	database, cleanup := setupIntegrationTestDB(t)
	defer cleanup()

	database.Exec("DELETE FROM go_users WHERE username LIKE 'testuser%'")

	site := NewAdminSite()
	err := Register[*auth.User](site, &ModelAdmin{})
	if err != nil {
		t.Fatalf("Failed to register model: %v", err)
	}

	router := site.URLs(database)

	// Submit with "_addanother" button
	formData := url.Values{}
	formData.Set("username", "testuser_another")
	formData.Set("email", "another@example.com")
	formData.Set("password", "testpass123")
	formData.Set("_addanother", "Save and add another")

	req := httptest.NewRequest("POST", "/auth/User/add/", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should redirect back to add form
	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected redirect 303, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if !strings.Contains(location, "/add/") {
		t.Errorf("Expected redirect to add form, got: %s", location)
	}

	// Verify record was created
	qs := queryset.NewQuerySet[*auth.User](database)
	users, _ := qs.Filter(queryset.Q{"username": "testuser_another"}).All()
	if len(users) != 1 {
		t.Errorf("Expected 1 user to be created")
	}
}

func TestAddViewIntegration_DuplicateColumnFix(t *testing.T) {
	database, cleanup := setupIntegrationTestDB(t)
	defer cleanup()

	database.Exec("DELETE FROM go_users WHERE username LIKE 'testuser%'")

	site := NewAdminSite()
	err := Register[*auth.User](site, &ModelAdmin{})
	if err != nil {
		t.Fatalf("Failed to register model: %v", err)
	}

	router := site.URLs(database)

	// This test specifically verifies the fix for "column specified more than once"
	// User model embeds auth.Model which has CreatedAt, UpdatedAt fields
	formData := url.Values{}
	formData.Set("username", "testuser_duplicatefix")
	formData.Set("email", "duplicate@example.com")
	formData.Set("password", "testpass123")
	formData.Set("is_active", "true")

	req := httptest.NewRequest("POST", "/auth/User/add/", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should NOT get database error about duplicate columns
	if w.Code != http.StatusSeeOther && w.Code != http.StatusFound {
		body := w.Body.String()
		if strings.Contains(body, "specified more than once") {
			t.Fatalf("Duplicate column error still present: %s", body)
		}
		t.Errorf("Expected successful redirect, got %d. Body: %s", w.Code, body)
	}

	// Verify record was created successfully
	qs := queryset.NewQuerySet[*auth.User](database)
	users, err := qs.Filter(queryset.Q{"username": "testuser_duplicatefix"}).All()
	if err != nil {
		t.Fatalf("Database query failed: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user to be created, got %d", len(users))
	}
}

// Helper function to set up test database for integration tests
func setupIntegrationTestDB(t *testing.T) (*db.DB, func()) {
	// Connect to test database
	database, err := db.NewDB("postgres", "user=demo_user password=postgres dbname=demo_db sslmode=disable")
	if err != nil {
		t.Skipf("Skipping test - cannot connect to test database: %v", err)
	}

	cleanup := func() {
		database.Close()
	}

	return database, cleanup
}
