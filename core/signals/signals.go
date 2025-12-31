package signals

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
)

// Handler processes signal emissions
type Handler[T any] func(sender any, data T) error

// handlerEntry stores a handler and its configuration
type handlerEntry[T any] struct {
	id      string
	handler Handler[T]
}

// Signal represents an event that can be emitted
type Signal[T any] struct {
	name     string
	handlers []handlerEntry[T]
	mu       sync.RWMutex
}

// NewSignal creates a named typed signal
func NewSignal[T any](name string) *Signal[T] {
	return &Signal[T]{
		name:     name,
		handlers: make([]handlerEntry[T], 0),
	}
}

// Connect registers a handler and returns a unique ID
func (s *Signal[T]) Connect(handler Handler[T]) string {
	id := generateID()
	_ = s.ConnectWithID(id, handler)
	return id
}

// ConnectWithID registers with a custom ID
func (s *Signal[T]) ConnectWithID(id string, handler Handler[T]) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for existing ID
	for _, entry := range s.handlers {
		if entry.id == id {
			return fmt.Errorf("handler with id %s already registered", id)
		}
	}

	s.handlers = append(s.handlers, handlerEntry[T]{
		id:      id,
		handler: handler,
	})
	return nil
}

// Disconnect removes a handler by ID
func (s *Signal[T]) Disconnect(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, entry := range s.handlers {
		if entry.id == id {
			s.handlers = append(s.handlers[:i], s.handlers[i+1:]...)
			return true
		}
	}
	return false
}

// Emit sends the signal to all handlers synchronously
func (s *Signal[T]) Emit(sender any, data T) []error {
	s.mu.RLock()
	// Copy handlers to avoid holding lock during execution
	handlers := append([]handlerEntry[T](nil), s.handlers...)
	s.mu.RUnlock()

	var errs []error
	for _, entry := range handlers {
		if err := entry.handler(sender, data); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// EmitAsync sends signal asynchronously in a new goroutine
func (s *Signal[T]) EmitAsync(sender any, data T) {
	go func() {
		_ = s.Emit(sender, data)
	}()
}

// HandlerCount returns number of connected handlers
func (s *Signal[T]) HandlerCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.handlers)
}

func generateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
