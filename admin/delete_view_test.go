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

// Integration tests for Delete view

func TestDeleteView_GET_ShowsConfirmation(t *testing.T) {
	database, cleanup := setupIntegrationTestDB(t)
	defer cleanup()

	// Create a test user
	user := &auth.User{
		Username: "deleteuser",
		Email:    "delete@example.com",
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

	// GET delete confirmation page
	req := httptest.NewRequest("GET", fmt.Sprintf("/auth/User/%d/delete/", user.ID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	body := w.Body.String()

	// Verify confirmation page elements
	if !strings.Contains(body, "Are you sure") {
		t.Error("Expected 'Are you sure' confirmation message")
	}

	if !strings.Contains(body, "deleteuser") {
		t.Error("Expected username to be displayed")
	}

	if !strings.Contains(body, "<form") {
		t.Error("Expected form element")
	}

	// Should have a submit button
	if !strings.Contains(body, "type=\"submit\"") {
		t.Error("Expected submit button")
	}

	// Should have a cancel link
	if !strings.Contains(body, "Cancel") {
		t.Error("Expected cancel button/link")
	}
}

func TestDeleteView_POST_DeletesRecord(t *testing.T) {
	database, cleanup := setupIntegrationTestDB(t)
	defer cleanup()

	// Create a test user
	user := &auth.User{
		Username: "todelete",
		Email:    "todelete@example.com",
		Password: "hashedpass",
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

	// POST to delete
	formData := url.Values{}
	req := httptest.NewRequest("POST", fmt.Sprintf("/auth/User/%d/delete/", userID), strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should redirect to list after deletion
	if w.Code != http.StatusSeeOther && w.Code != http.StatusFound {
		t.Fatalf("Expected redirect (303/302), got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if !strings.Contains(location, "/auth/User/") {
		t.Errorf("Expected redirect to list view, got: %s", location)
	}

	// Verify record was actually deleted
	qs2 := queryset.NewQuerySet[*auth.User](database)
	_, err = qs2.GetByID(userID)
	if err == nil {
		t.Error("Expected error when fetching deleted user, but got none")
	}
}

func TestDeleteView_NotFound(t *testing.T) {
	database, cleanup := setupIntegrationTestDB(t)
	defer cleanup()

	site := NewAdminSite()
	err := Register[*auth.User](site, &ModelAdmin{})
	if err != nil {
		t.Fatalf("Failed to register model: %v", err)
	}

	router := site.URLs(database)

	// Try to delete non-existent ID
	req := httptest.NewRequest("GET", "/auth/User/99999/delete/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 404
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for non-existent ID, got %d", w.Code)
	}
}

func TestDeleteView_RedirectsToList(t *testing.T) {
	database, cleanup := setupIntegrationTestDB(t)
	defer cleanup()

	// Create a test user
	user := &auth.User{
		Username: "redirecttest",
		Email:    "redirect@example.com",
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

	// DELETE the user
	req := httptest.NewRequest("POST", fmt.Sprintf("/auth/User/%d/delete/", user.ID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Check redirect location
	location := w.Header().Get("Location")
	expectedLocation := "/admin/auth/User/"

	if !strings.HasSuffix(location, expectedLocation) {
		t.Errorf("Expected redirect to %s, got: %s", expectedLocation, location)
	}
}
