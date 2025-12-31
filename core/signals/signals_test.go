package signals

import (
	"errors"
	"testing"
	"time"
)

func TestSignal(t *testing.T) {
	t.Run("connects and emits signal", func(t *testing.T) {
		type TestData struct{ Message string }
		s := NewSignal[TestData]("test_signal")

		var received string
		handler := func(sender any, data TestData) error {
			received = data.Message
			return nil
		}

		s.Connect(handler)

		errs := s.Emit("sender1", TestData{Message: "hello"})

		if len(errs) != 0 {
			t.Errorf("expected 0 errors, got %d", len(errs))
		}
		if received != "hello" {
			t.Errorf("expected 'hello', got '%s'", received)
		}
	})

	t.Run("handles multiple handlers", func(t *testing.T) {
		type TestData struct{ Count int }
		s := NewSignal[TestData]("test_count")

		total := 0
		s.Connect(func(sender any, data TestData) error {
			total += data.Count
			return nil
		})
		s.Connect(func(sender any, data TestData) error {
			total += data.Count * 2
			return nil
		})

		s.Emit(nil, TestData{Count: 1})

		if total != 3 {
			t.Errorf("expected 3, got %d", total)
		}
	})

	t.Run("disconnects handler by ID", func(t *testing.T) {
		type TestData struct{}
		s := NewSignal[TestData]("test_disconnect")

		called := false
		err := s.ConnectWithID("handler1", func(sender any, data TestData) error {
			called = true
			return nil
		})
		if err != nil {
			t.Fatalf("ConnectWithID failed: %v", err)
		}
		id := "handler1"

		s.Disconnect(id)
		s.Emit(nil, TestData{})

		if called {
			t.Error("handler was called after disconnection")
		}
	})

	t.Run("collects multiple errors", func(t *testing.T) {
		type TestData struct{}
		s := NewSignal[TestData]("test_errors")

		s.Connect(func(sender any, data TestData) error {
			return errors.New("error 1")
		})
		s.Connect(func(sender any, data TestData) error {
			return errors.New("error 2")
		})

		errs := s.Emit(nil, TestData{})

		if len(errs) != 2 {
			t.Errorf("expected 2 errors, got %d", len(errs))
		}
	})

	t.Run("emits asynchronously", func(t *testing.T) {
		type TestData struct{ Done chan bool }
		s := NewSignal[TestData]("test_async")

		s.Connect(func(sender any, data TestData) error {
			time.Sleep(10 * time.Millisecond)
			data.Done <- true
			return nil
		})

		done := make(chan bool)
		s.EmitAsync(nil, TestData{Done: done})

		select {
		case <-done:
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Error("timeout waiting for async signal")
		}
	})
}
