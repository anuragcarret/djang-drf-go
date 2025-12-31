package apps

import (
	"errors"
	"testing"
)

// MockApp is a test implementation of the App interface
type MockApp struct {
	config      *AppConfig
	readyCalled bool
	readyError  error
}

func (m *MockApp) AppConfig() *AppConfig {
	return m.config
}

func (m *MockApp) Ready() error {
	m.readyCalled = true
	return m.readyError
}

// Helper to create a mock app
func newMockApp(name string) *MockApp {
	return &MockApp{
		config: &AppConfig{
			Name:        name,
			VerboseName: name + " App",
		},
	}
}

// =============================================================================
// App Interface Tests
// =============================================================================

func TestAppInterface(t *testing.T) {
	t.Run("AppConfig returns configuration", func(t *testing.T) {
		app := newMockApp("users")

		config := app.AppConfig()

		if config == nil {
			t.Fatal("AppConfig() returned nil")
		}
		if config.Name != "users" {
			t.Errorf("expected Name 'users', got '%s'", config.Name)
		}
	})

	t.Run("Ready is called by registry", func(t *testing.T) {
		app := newMockApp("users")

		if app.readyCalled {
			t.Error("Ready() should not be called before Populate()")
		}

		_ = app.Ready()

		if !app.readyCalled {
			t.Error("Ready() was not called")
		}
	})
}

// =============================================================================
// Registry Tests
// =============================================================================

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()

	if r == nil {
		t.Fatal("NewRegistry() returned nil")
	}
	if r.apps == nil {
		t.Error("apps map not initialized")
	}
	if r.IsReady() {
		t.Error("new registry should not be ready")
	}
	if r.AppCount() != 0 {
		t.Errorf("expected 0 apps, got %d", r.AppCount())
	}
}

func TestRegistryRegister(t *testing.T) {
	t.Run("registers app successfully", func(t *testing.T) {
		r := NewRegistry()
		app := newMockApp("users")

		err := r.Register(app)

		if err != nil {
			t.Fatalf("Register() failed: %v", err)
		}
		if r.AppCount() != 1 {
			t.Errorf("expected 1 app, got %d", r.AppCount())
		}
	})

	t.Run("rejects duplicate app names", func(t *testing.T) {
		r := NewRegistry()
		app1 := newMockApp("users")
		app2 := newMockApp("users")

		_ = r.Register(app1)
		err := r.Register(app2)

		if !errors.Is(err, ErrDuplicateApp) {
			t.Errorf("expected ErrDuplicateApp, got %v", err)
		}
	})

	t.Run("rejects empty app name", func(t *testing.T) {
		r := NewRegistry()
		app := &MockApp{config: &AppConfig{Name: ""}}

		err := r.Register(app)

		if !errors.Is(err, ErrEmptyAppName) {
			t.Errorf("expected ErrEmptyAppName, got %v", err)
		}
	})

	t.Run("rejects nil AppConfig", func(t *testing.T) {
		r := NewRegistry()
		app := &MockApp{config: nil}

		err := r.Register(app)

		if !errors.Is(err, ErrEmptyAppName) {
			t.Errorf("expected ErrEmptyAppName, got %v", err)
		}
	})

	t.Run("rejects registration after Populate", func(t *testing.T) {
		r := NewRegistry()
		app1 := newMockApp("users")
		app2 := newMockApp("orders")

		_ = r.Register(app1)
		_ = r.Populate()
		err := r.Register(app2)

		if !errors.Is(err, ErrRegistryPopulated) {
			t.Errorf("expected ErrRegistryPopulated, got %v", err)
		}
	})
}

func TestRegistryPopulate(t *testing.T) {
	t.Run("calls Ready on all apps", func(t *testing.T) {
		r := NewRegistry()
		app1 := newMockApp("users")
		app2 := newMockApp("orders")

		_ = r.Register(app1)
		_ = r.Register(app2)

		err := r.Populate()

		if err != nil {
			t.Fatalf("Populate() failed: %v", err)
		}
		if !app1.readyCalled {
			t.Error("app1.Ready() was not called")
		}
		if !app2.readyCalled {
			t.Error("app2.Ready() was not called")
		}
	})

	t.Run("marks registry as ready", func(t *testing.T) {
		r := NewRegistry()
		app := newMockApp("users")

		_ = r.Register(app)
		_ = r.Populate()

		if !r.IsReady() {
			t.Error("registry should be ready after Populate()")
		}
	})

	t.Run("stops on Ready error", func(t *testing.T) {
		r := NewRegistry()
		app1 := newMockApp("users")
		app2 := newMockApp("orders")
		app2.readyError = errors.New("initialization failed")

		_ = r.Register(app1)
		_ = r.Register(app2)

		err := r.Populate()

		if err == nil {
			t.Error("expected error from failing Ready()")
		}
	})

	t.Run("can be called multiple times safely", func(t *testing.T) {
		r := NewRegistry()
		app := newMockApp("users")

		_ = r.Register(app)
		_ = r.Populate()
		err := r.Populate()

		if err != nil {
			t.Errorf("second Populate() should not fail: %v", err)
		}
	})
}

func TestRegistryGetApp(t *testing.T) {
	t.Run("returns registered app", func(t *testing.T) {
		r := NewRegistry()
		app := newMockApp("users")
		_ = r.Register(app)

		result, err := r.GetApp("users")

		if err != nil {
			t.Fatalf("GetApp() failed: %v", err)
		}
		if result != app {
			t.Error("returned app does not match registered app")
		}
	})

	t.Run("returns error for unknown app", func(t *testing.T) {
		r := NewRegistry()

		_, err := r.GetApp("unknown")

		if !errors.Is(err, ErrAppNotFound) {
			t.Errorf("expected ErrAppNotFound, got %v", err)
		}
	})
}

func TestRegistryGetApps(t *testing.T) {
	t.Run("returns all apps in order", func(t *testing.T) {
		r := NewRegistry()
		app1 := newMockApp("users")
		app2 := newMockApp("orders")
		app3 := newMockApp("products")

		_ = r.Register(app1)
		_ = r.Register(app2)
		_ = r.Register(app3)

		apps := r.GetApps()

		if len(apps) != 3 {
			t.Fatalf("expected 3 apps, got %d", len(apps))
		}
		if apps[0].AppConfig().Name != "users" {
			t.Error("first app should be 'users'")
		}
		if apps[1].AppConfig().Name != "orders" {
			t.Error("second app should be 'orders'")
		}
		if apps[2].AppConfig().Name != "products" {
			t.Error("third app should be 'products'")
		}
	})

	t.Run("returns empty slice for empty registry", func(t *testing.T) {
		r := NewRegistry()

		apps := r.GetApps()

		if len(apps) != 0 {
			t.Errorf("expected 0 apps, got %d", len(apps))
		}
	})
}

func TestRegistryGetAppConfigs(t *testing.T) {
	r := NewRegistry()
	app := newMockApp("users")
	_ = r.Register(app)

	configs := r.GetAppConfigs()

	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}
	if configs[0].Name != "users" {
		t.Errorf("expected 'users', got '%s'", configs[0].Name)
	}
}

// =============================================================================
// Global Registry Tests
// =============================================================================

func TestGlobalRegistry(t *testing.T) {
	if Apps == nil {
		t.Fatal("global Apps registry should be initialized")
	}
}

// =============================================================================
// Concurrency Tests
// =============================================================================

func TestRegistryConcurrency(t *testing.T) {
	t.Run("concurrent reads are safe", func(t *testing.T) {
		r := NewRegistry()
		app := newMockApp("users")
		_ = r.Register(app)
		_ = r.Populate()

		done := make(chan bool)
		for i := 0; i < 100; i++ {
			go func() {
				_, _ = r.GetApp("users")
				_ = r.GetApps()
				_ = r.IsReady()
				done <- true
			}()
		}

		for i := 0; i < 100; i++ {
			<-done
		}
	})
}
