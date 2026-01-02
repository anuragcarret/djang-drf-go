package admin

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/anuragcarret/djang-drf-go/contrib/auth"
	"github.com/anuragcarret/djang-drf-go/orm/queryset"
)

// Integration tests for Change/Edit view with real database operations

func TestChangeViewIntegration_GetForm(t *testing.T) {
	database, cleanup := setupIntegrationTestDB(t)
	defer cleanup()

	// Create a test user first
	user := &auth.User{
		Username:  "edituser",
		Email:     "edituser@example.com",
		Password:  "hashedpass",
		FirstName: "Edit",
		LastName:  "User",
		IsActive:  true,
	}

	qs := queryset.NewQuerySet[*auth.User](database)
	err := qs.Create(user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Register User model
	site := NewAdminSite()
	err = Register[*auth.User](site, &ModelAdmin{
		ListDisplay: []string{"ID", "Username", "Email"},
	})
	if err != nil {
		t.Fatalf("Failed to register model: %v", err)
	}

	router := site.URLs(database)

	// GET the change form
	req := httptest.NewRequest("GET", fmt.Sprintf("/auth/User/%d/change/", user.ID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	body := w.Body.String()

	// Verify form is rendered
	if !strings.Contains(body, "<form") {
		t.Error("Expected form element in response")
	}

	// Verify form is pre-filled with user data
	if !strings.Contains(body, "edituser") {
		t.Error("Expected username 'edituser' to be pre-filled")
	}

	if !strings.Contains(body, "edituser@example.com") {
		t.Error("Expected email to be pre-filled")
	}

	if !strings.Contains(body, "Edit") {
		t.Error("Expected first name to be pre-filled")
	}

	// Verify delete button is present
	if !strings.Contains(body, "Delete") {
		t.Error("Expected Delete button in edit form")
	}

	// Verify form action points to change URL
	if !strings.Contains(body, fmt.Sprintf("/change/")) {
		t.Error("Expected form action to point to change URL")
	}
}

func TestChangeViewIntegration_UpdateRecord(t *testing.T) {
	database, cleanup := setupIntegrationTestDB(t)
	defer cleanup()

	// Create a test user first
	user := &auth.User{
		Username:  "updateuser",
		Email:     "before@example.com",
		Password:  "hashedpass",
		FirstName: "Before",
		IsActive:  false,
	}

	qs := queryset.NewQuerySet[*auth.User](database)
	err := qs.Create(user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	userID := user.ID

	site := NewAdminSite()
	err = Register[*auth.User](site, &ModelAdmin{})
	if err != nil {
		t.Fatalf("Failed to register model: %v", err)
	}

	router := site.URLs(database)

	// Update the user via POST
	formData := url.Values{}
	formData.Set("username", "updateuser")     // Keep same
	formData.Set("email", "after@example.com") // Changed
	formData.Set("password", "hashedpass")
	formData.Set("first_name", "After")  // Changed
	formData.Set("last_name", "Updated") // Changed
	formData.Set("is_active", "true")    // Changed

	req := httptest.NewRequest("POST", fmt.Sprintf("/auth/User/%d/change/", userID), strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should redirect after successful update
	if w.Code != http.StatusFound && w.Code != http.StatusSeeOther {
		t.Fatalf("Expected redirect (302/303), got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify record was updated in database
	qs2 := queryset.NewQuerySet[*auth.User](database)
	updatedUser, err := qs2.GetByID(userID)
	if err != nil {
		t.Fatalf("Failed to fetch updated user: %v", err)
	}

	if updatedUser.Email != "after@example.com" {
		t.Errorf("Expected email 'after@example.com', got '%s'", updatedUser.Email)
	}

	if updatedUser.FirstName != "After" {
		t.Errorf("Expected first_name 'After', got '%s'", updatedUser.FirstName)
	}

	if updatedUser.LastName != "Updated" {
		t.Errorf("Expected last_name 'Updated', got '%s'", updatedUser.LastName)
	}

	if !updatedUser.IsActive {
		t.Error("Expected is_active to be true")
	}
}

func TestChangeViewIntegration_NotFound(t *testing.T) {
	database, cleanup := setupIntegrationTestDB(t)
	defer cleanup()

	site := NewAdminSite()
	err := Register[*auth.User](site, &ModelAdmin{})
	if err != nil {
		t.Fatalf("Failed to register model: %v", err)
	}

	router := site.URLs(database)

	// Try to access non-existent ID
	req := httptest.NewRequest("GET", "/auth/User/99999/change/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 404 or error message
	if w.Code != http.StatusNotFound && w.Code != http.StatusOK {
		// If 200, should contain error message
		body := w.Body.String()
		if !strings.Contains(body, "not found") && !strings.Contains(body, "does not exist") {
			t.Errorf("Expected 404 or error message for non-existent ID, got %d", w.Code)
		}
	}
}

func TestChangeViewIntegration_SaveAndContinue(t *testing.T) {
	database, cleanup := setupIntegrationTestDB(t)
	defer cleanup()

	// Create test user
	user := &auth.User{
		Username: "continueuser",
		Email:    "continue@example.com",
		Password: "hashedpass",
	}

	qs := queryset.NewQuerySet[*auth.User](database)
	err := qs.Create(user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	site := NewAdminSite()
	err = Register[*auth.User](site, &ModelAdmin{})
	if err != nil {
		t.Fatalf("Failed to register model: %v", err)
	}

	router := site.URLs(database)

	// Update with "_continue" button
	formData := url.Values{}
	formData.Set("username", "continueuser")
	formData.Set("email", "continue_updated@example.com")
	formData.Set("password", "hashedpass")
	formData.Set("_continue", "Save and continue editing")

	req := httptest.NewRequest("POST", fmt.Sprintf("/auth/User/%d/change/", user.ID), strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should redirect back to change form
	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected redirect 303, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if !strings.Contains(location, "/change/") {
		t.Errorf("Expected redirect to change form, got: %s", location)
	}

	// Verify update was saved
	qs2 := queryset.NewQuerySet[*auth.User](database)
	updatedUser, _ := qs2.GetByID(user.ID)
	if updatedUser.Email != "continue_updated@example.com" {
		t.Errorf("Expected email to be updated")
	}
}

func TestChangeViewIntegration_ValidationErrors(t *testing.T) {
	database, cleanup := setupIntegrationTestDB(t)
	defer cleanup()

	// Create test user
	user := &auth.User{
		Username: "validationuser",
		Email:    "valid@example.com",
		Password: "hashedpass",
	}

	qs := queryset.NewQuerySet[*auth.User](database)
	err := qs.Create(user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	site := NewAdminSite()
	err = Register[*auth.User](site, &ModelAdmin{})
	if err != nil {
		t.Fatalf("Failed to register model: %v", err)
	}

	router := site.URLs(database)

	// Update with missing required field (username)
	formData := url.Values{}
	formData.Set("username", "") // Empty username
	formData.Set("email", "valid@example.com")
	formData.Set("password", "hashedpass")

	req := httptest.NewRequest("POST", fmt.Sprintf("/auth/User/%d/change/", user.ID), strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 200 with error (not redirect)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for validation error, got %d", w.Code)
	}

	body := w.Body.String()

	// Verify error message
	if !strings.Contains(body, "required") {
		t.Error("Expected 'required' error message")
	}

	// Verify form is re-rendered
	if !strings.Contains(body, "<form") {
		t.Error("Expected form to be re-rendered")
	}
}
