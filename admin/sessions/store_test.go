package sessions

import (
	"testing"
	"time"
)

func TestInMemoryStore_SetGet(t *testing.T) {
	store := NewInMemorySessionStore()

	session := &SessionData{
		UserID:      123,
		Username:    "testuser",
		IsStaff:     true,
		IsSuperuser: false,
		CreatedAt:   time.Now(),
		LastAccess:  time.Now(),
		CSRFToken:   "test-csrf-token",
	}

	// Set session
	err := store.Set("session123", session, 1*time.Hour)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get session
	retrieved, err := store.Get("session123")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Retrieved session is nil")
	}

	if retrieved.UserID != 123 {
		t.Errorf("Expected UserID=123, got %d", retrieved.UserID)
	}

	if retrieved.Username != "testuser" {
		t.Errorf("Expected Username=testuser, got %s", retrieved.Username)
	}

	if !retrieved.IsStaff {
		t.Error("Expected IsStaff=true")
	}
}

func TestInMemoryStore_GetNonExistent(t *testing.T) {
	store := NewInMemorySessionStore()

	session, err := store.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent session, got nil")
	}

	if session != nil {
		t.Error("Expected nil session for non-existent ID")
	}
}

func TestInMemoryStore_Delete(t *testing.T) {
	store := NewInMemorySessionStore()

	session := &SessionData{
		UserID:   123,
		Username: "testuser",
	}

	store.Set("session123", session, 1*time.Hour)

	// Delete
	err := store.Delete("session123")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	_, err = store.Get("session123")
	if err == nil {
		t.Error("Expected error after deletion, got nil")
	}
}

func TestInMemoryStore_Expiry(t *testing.T) {
	store := NewInMemorySessionStore()

	session := &SessionData{
		UserID:   123,
		Username: "testuser",
	}

	// Set with very short expiry
	store.Set("session123", session, 100*time.Millisecond)

	// Should exist immediately
	retrieved, err := store.Get("session123")
	if err != nil {
		t.Error("Session should exist immediately after creation")
	}
	if retrieved == nil {
		t.Error("Retrieved session should not be nil")
	}

	// Wait for expiry
	time.Sleep(200 * time.Millisecond)

	// Should be expired
	_, err = store.Get("session123")
	if err == nil {
		t.Error("Expected error for expired session, got nil")
	}
}

func TestInMemoryStore_Cleanup(t *testing.T) {
	store := NewInMemorySessionStore()

	// Add expired session
	session1 := &SessionData{UserID: 1, Username: "expired"}
	store.Set("expired", session1, 1*time.Millisecond)

	// Add valid session
	session2 := &SessionData{UserID: 2, Username: "valid"}
	store.Set("valid", session2, 1*time.Hour)

	// Wait for first to expire
	time.Sleep(10 * time.Millisecond)

	// Run cleanup
	err := store.Cleanup()
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Expired should be gone
	_, err = store.Get("expired")
	if err == nil {
		t.Error("Expired session should have been cleaned up")
	}

	// Valid should still exist
	retrieved, err := store.Get("valid")
	if err != nil {
		t.Error("Valid session should still exist after cleanup")
	}
	if retrieved == nil {
		t.Error("Retrieved valid session should not be nil")
	}
}

func TestInMemoryStore_UpdateSession(t *testing.T) {
	store := NewInMemorySessionStore()

	session := &SessionData{
		UserID:     123,
		Username:   "testuser",
		LastAccess: time.Now(),
	}

	store.Set("session123", session, 1*time.Hour)

	// Update last access time
	time.Sleep(10 * time.Millisecond)
	session.LastAccess = time.Now()
	store.Set("session123", session, 1*time.Hour)

	// Retrieve and verify
	retrieved, _ := store.Get("session123")

	// Last access should be updated
	if retrieved.LastAccess.Before(session.LastAccess) {
		t.Error("Last access time should have been updated")
	}
}
