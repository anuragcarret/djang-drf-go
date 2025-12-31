package management

import (
	"context"
	"errors"
	"testing"
)

// MockCommand is a test implementation of Command
type MockCommand struct {
	name    string
	runFunc func(ctx context.Context, args []string) error
	called  bool
}

func (c *MockCommand) Name() string { return c.name }
func (c *MockCommand) Help() string { return "Mock command help" }
func (c *MockCommand) Run(ctx context.Context, args []string) error {
	c.called = true
	if c.runFunc != nil {
		return c.runFunc(ctx, args)
	}
	return nil
}

func TestRegistry(t *testing.T) {
	t.Run("registers and retrieves command", func(t *testing.T) {
		r := NewRegistry()
		cmd := &MockCommand{name: "test"}
		r.Register(cmd)

		result, err := r.Get("test")
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}
		if result != cmd {
			t.Error("retrieved command mismatch")
		}
	})

	t.Run("returns error for unknown command", func(t *testing.T) {
		r := NewRegistry()
		_, err := r.Get("unknown")
		if !errors.Is(err, ErrCommandNotFound) {
			t.Errorf("expected ErrCommandNotFound, got %v", err)
		}
	})
}

func TestExecute(t *testing.T) {
	t.Run("executes registered command", func(t *testing.T) {
		r := NewRegistry()
		cmd := &MockCommand{name: "hello"}
		r.Register(cmd)

		err := r.Execute(context.Background(), []string{"hello"})
		if err != nil {
			t.Fatalf("Execute() failed: %v", err)
		}
		if !cmd.called {
			t.Error("command was not executed")
		}
	})

	t.Run("returns error for empty args", func(t *testing.T) {
		r := NewRegistry()
		err := r.Execute(context.Background(), []string{})
		if err == nil {
			t.Error("expected error for empty args")
		}
	})
}

func TestFlagParsing(t *testing.T) {
	t.Run("binds flags from struct tags", func(t *testing.T) {
		type TestCmd struct {
			App  string `flag:"app" default:"core"`
			Fake bool   `flag:"fake"`
		}
		cmd := &TestCmd{}
		fs := BindFlags(cmd)

		err := fs.Parse([]string{"--app", "users", "--fake"})
		if err != nil {
			t.Fatalf("Parse() failed: %v", err)
		}

		if cmd.App != "users" {
			t.Errorf("expected 'users', got '%s'", cmd.App)
		}
		if !cmd.Fake {
			t.Error("expected Fake to be true")
		}
	})

	t.Run("uses default values", func(t *testing.T) {
		type TestCmd struct {
			Port int `flag:"port" default:"8000"`
		}
		cmd := &TestCmd{}
		fs := BindFlags(cmd)

		_ = fs.Parse([]string{})

		if cmd.Port != 8000 {
			t.Errorf("expected 8000, got %d", cmd.Port)
		}
	})
}
