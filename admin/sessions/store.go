package sessions

import (
	"errors"
	"sync"
	"time"
)

// SessionData holds information about an admin session
type SessionData struct {
	UserID      uint64
	Username    string
	IsStaff     bool
	IsSuperuser bool
	CreatedAt   time.Time
	LastAccess  time.Time
	CSRFToken   string
}

// SessionStore defines the interface for session storage
type SessionStore interface {
	Set(sessionID string, data *SessionData, expiry time.Duration) error
	Get(sessionID string) (*SessionData, error)
	Delete(sessionID string) error
	Cleanup() error
}

// InMemorySessionStore implements SessionStore using in-memory storage
type InMemorySessionStore struct {
	sessions map[string]*sessionEntry
	mu       sync.RWMutex
}

type sessionEntry struct {
	data      *SessionData
	expiresAt time.Time
}

// NewInMemorySessionStore creates a new in-memory session store
func NewInMemorySessionStore() *InMemorySessionStore {
	store := &InMemorySessionStore{
		sessions: make(map[string]*sessionEntry),
	}

	// Start cleanup goroutine
	go store.cleanupLoop()

	return store
}

// Set stores a session with the given expiry duration
func (s *InMemorySessionStore) Set(sessionID string, data *SessionData, expiry time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[sessionID] = &sessionEntry{
		data:      data,
		expiresAt: time.Now().Add(expiry),
	}

	return nil
}

// Get retrieves a session by ID
func (s *InMemorySessionStore) Get(sessionID string) (*SessionData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.sessions[sessionID]
	if !exists {
		return nil, errors.New("session not found")
	}

	// Check if expired
	if time.Now().After(entry.expiresAt) {
		return nil, errors.New("session expired")
	}

	return entry.data, nil
}

// Delete removes a session
func (s *InMemorySessionStore) Delete(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, sessionID)
	return nil
}

// Cleanup removes all expired sessions
func (s *InMemorySessionStore) Cleanup() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for sessionID, entry := range s.sessions {
		if now.After(entry.expiresAt) {
			delete(s.sessions, sessionID)
		}
	}

	return nil
}

// cleanupLoop periodically cleans up expired sessions
func (s *InMemorySessionStore) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.Cleanup()
	}
}
