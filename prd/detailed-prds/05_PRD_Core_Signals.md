# PRD: Core Signals System

> **Module:** `core/signals`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **Django Equivalent:** `django.dispatch`

---

## 1. Overview

### 1.1 Purpose

The Signals System enables decoupled event-driven communication:

- **Publisher/Subscriber Pattern**: Components communicate without tight coupling
- **Lifecycle Events**: Hook into model saves, deletes, request lifecycle
- **Custom Signals**: Define application-specific events
- **Async Support**: Non-blocking signal handlers

### 1.2 Design Philosophy

| Django Approach | Go-Idiomatic Approach |
|-----------------|----------------------|
| `Signal` class with `send()` | Generic `Signal[T]` with `Emit()` |
| `@receiver` decorator | `signal.Connect(handler)` |
| `dispatch_uid` for deduplication | Handler registration with ID |
| Synchronous by default | Options for sync/async dispatch |

---

## 2. Core Concepts

### 2.1 Signal Definition

```go
// Signal represents an event that can be emitted
type Signal[T any] struct {
    name     string
    handlers []Handler[T]
    mu       sync.RWMutex
}

// Handler processes signal emissions
type Handler[T any] func(sender any, data T) error

// NewSignal creates a named signal
func NewSignal[T any](name string) *Signal[T]
```

### 2.2 Built-in Signals

#### Model Signals

```go
var (
    // PreSave fires before model save
    PreSave = NewSignal[ModelSignalData]("pre_save")
    
    // PostSave fires after model save
    PostSave = NewSignal[ModelSignalData]("post_save")
    
    // PreDelete fires before model delete
    PreDelete = NewSignal[ModelSignalData]("pre_delete")
    
    // PostDelete fires after model delete
    PostDelete = NewSignal[ModelSignalData]("post_delete")
    
    // M2MChanged fires on many-to-many changes
    M2MChanged = NewSignal[M2MSignalData]("m2m_changed")
)

// ModelSignalData carries model signal information
type ModelSignalData struct {
    Instance interface{}  // The model instance
    Created  bool         // True if new record (PostSave only)
    Update   bool         // True if update (PostSave only)
    Fields   []string     // Fields being updated
    Raw      bool         // True for raw SQL operations
    Using    string       // Database alias
}
```

#### Request Signals

```go
var (
    // RequestStarted fires at request start
    RequestStarted = NewSignal[RequestSignalData]("request_started")
    
    // RequestFinished fires at request end
    RequestFinished = NewSignal[RequestSignalData]("request_finished")
    
    // GotRequestException fires on unhandled exception
    GotRequestException = NewSignal[ExceptionSignalData]("got_request_exception")
)

// RequestSignalData carries request information
type RequestSignalData struct {
    Request   *http.Request
    TraceID   string
    StartTime time.Time
}

// ExceptionSignalData for exception signals
type ExceptionSignalData struct {
    Request   *http.Request
    Exception error
    TraceID   string
}
```

#### System Signals

```go
var (
    // AppReady fires when an app is ready
    AppReady = NewSignal[AppReadyData]("app_ready")
    
    // SettingChanged fires on settings modification
    SettingChanged = NewSignal[SettingChangedData]("setting_changed")
    
    // MigrationApplied fires after migration runs
    MigrationApplied = NewSignal[MigrationData]("migration_applied")
)
```

---

## 3. API Specification

### 3.1 Signal Definition

```go
// NewSignal creates a typed signal
func NewSignal[T any](name string) *Signal[T]

// Example: Custom signal definition
type UserRegisteredData struct {
    User      *User
    IPAddress string
    UserAgent string
}

var UserRegistered = signals.NewSignal[UserRegisteredData]("user_registered")
```

### 3.2 Connecting Handlers

```go
// Connect registers a handler
func (s *Signal[T]) Connect(handler Handler[T]) string

// ConnectWithID registers with custom ID for later disconnection
func (s *Signal[T]) ConnectWithID(id string, handler Handler[T]) error

// Disconnect removes a handler by ID
func (s *Signal[T]) Disconnect(id string) bool

// DisconnectAll removes all handlers
func (s *Signal[T]) DisconnectAll()
```

### 3.3 Emitting Signals

```go
// Emit sends the signal to all handlers
func (s *Signal[T]) Emit(sender any, data T) []error

// EmitAsync sends signal asynchronously
func (s *Signal[T]) EmitAsync(sender any, data T)

// EmitWithTimeout sends with timeout per handler
func (s *Signal[T]) EmitWithTimeout(sender any, data T, timeout time.Duration) []error
```

### 3.4 Query Handlers

```go
// HasHandlers checks if any handlers are connected
func (s *Signal[T]) HasHandlers() bool

// HandlerCount returns number of connected handlers
func (s *Signal[T]) HandlerCount() int

// HandlerIDs returns all registered handler IDs
func (s *Signal[T]) HandlerIDs() []string
```

---

## 4. Usage Examples

### 4.1 Basic Signal Usage

```go
// Define signal in signals.go
package users

import "github.com/yourusername/django-drf-go/core/signals"

type UserCreatedData struct {
    User *User
}

var UserCreated = signals.NewSignal[UserCreatedData]("user_created")

// Connect handler in app Ready()
func (a *UsersApp) Ready() error {
    UserCreated.Connect(onUserCreated)
    return nil
}

func onUserCreated(sender any, data UserCreatedData) error {
    // Send welcome email
    return email.SendWelcome(data.User.Email)
}

// Emit signal in views/services
func CreateUser(data CreateUserRequest) (*User, error) {
    user := &User{Email: data.Email}
    if err := user.Save(); err != nil {
        return nil, err
    }
    
    // Emit signal
    UserCreated.Emit(nil, UserCreatedData{User: user})
    
    return user, nil
}
```

### 4.2 Model Signal Handlers

```go
// Connect to built-in model signals
func (a *AuditApp) Ready() error {
    // Log all model saves
    signals.PostSave.Connect(logModelChange)
    
    // Validate before delete
    signals.PreDelete.Connect(validateDeletion)
    
    return nil
}

func logModelChange(sender any, data signals.ModelSignalData) error {
    action := "updated"
    if data.Created {
        action = "created"
    }
    
    log.Printf("Model %T %s by %v", data.Instance, action, sender)
    return nil
}

func validateDeletion(sender any, data signals.ModelSignalData) error {
    // Prevent deletion of protected records
    if user, ok := data.Instance.(*User); ok {
        if user.IsProtected {
            return errors.New("cannot delete protected user")
        }
    }
    return nil
}
```

### 4.3 Request Signals

```go
// Track request metrics
func (a *MetricsApp) Ready() error {
    signals.RequestStarted.Connect(recordRequestStart)
    signals.RequestFinished.Connect(recordRequestEnd)
    return nil
}

func recordRequestStart(sender any, data signals.RequestSignalData) error {
    metrics.IncrementCounter("http_requests_active")
    return nil
}

func recordRequestEnd(sender any, data signals.RequestSignalData) error {
    metrics.DecrementCounter("http_requests_active")
    duration := time.Since(data.StartTime)
    metrics.RecordHistogram("http_request_duration", duration.Seconds())
    return nil
}
```

---

## 5. Django Comparison

### 5.1 Signal Definition

**Django:**
```python
from django.dispatch import Signal

user_registered = Signal()  # providing_args removed in Django 3.1
```

**Django-DRF-Go:**
```go
type UserRegisteredData struct {
    User      *User
    IPAddress string
}

var UserRegistered = signals.NewSignal[UserRegisteredData]("user_registered")
```

### 5.2 Connecting Handlers

**Django:**
```python
from django.dispatch import receiver
from myapp.signals import user_registered

@receiver(user_registered)
def on_user_registered(sender, **kwargs):
    user = kwargs['user']
    send_welcome_email(user.email)

# Or manual connection
user_registered.connect(on_user_registered)
```

**Django-DRF-Go:**
```go
import "myproject/signals"

func init() {
    signals.UserRegistered.Connect(onUserRegistered)
}

func onUserRegistered(sender any, data signals.UserRegisteredData) error {
    return sendWelcomeEmail(data.User.Email)
}
```

### 5.3 Emitting Signals

**Django:**
```python
user_registered.send(sender=self.__class__, user=user, ip=request.META['REMOTE_ADDR'])
```

**Django-DRF-Go:**
```go
signals.UserRegistered.Emit(user, UserRegisteredData{
    User:      user,
    IPAddress: request.RemoteAddr,
})
```

### 5.4 Model Signals

**Django:**
```python
from django.db.models.signals import post_save
from django.dispatch import receiver
from .models import User

@receiver(post_save, sender=User)
def user_saved(sender, instance, created, **kwargs):
    if created:
        print(f"New user: {instance.email}")
```

**Django-DRF-Go:**
```go
func (a *MyApp) Ready() error {
    signals.PostSave.Connect(func(sender any, data signals.ModelSignalData) error {
        user, ok := data.Instance.(*User)
        if !ok {
            return nil  // Not a User model
        }
        if data.Created {
            fmt.Printf("New user: %s\n", user.Email)
        }
        return nil
    })
    return nil
}
```

---

## 6. Advanced Features

### 6.1 Sender Filtering

```go
// ConnectForSender only triggers for specific sender types
func (s *Signal[T]) ConnectForSender(sender reflect.Type, handler Handler[T]) string

// Example: Only listen to User model saves
signals.PostSave.ConnectForSender(
    reflect.TypeOf(&User{}),
    func(sender any, data signals.ModelSignalData) error {
        user := data.Instance.(*User)
        // Handle user save
        return nil
    },
)
```

### 6.2 Priority Handlers

```go
// ConnectWithPriority sets handler execution order
func (s *Signal[T]) ConnectWithPriority(priority int, handler Handler[T]) string

// Lower priority = earlier execution
signals.PreSave.ConnectWithPriority(1, validateHandler)   // Runs first
signals.PreSave.ConnectWithPriority(10, cacheHandler)     // Runs second
signals.PreSave.ConnectWithPriority(100, loggingHandler)  // Runs last
```

### 6.3 Async Dispatch

```go
// AsyncSignal for non-blocking handlers
type AsyncSignal[T any] struct {
    Signal[T]
    workers int
    queue   chan emission[T]
}

func NewAsyncSignal[T any](name string, workers int) *AsyncSignal[T]

// Start begins processing emissions
func (s *AsyncSignal[T]) Start(ctx context.Context)

// Usage
var UserActivity = signals.NewAsyncSignal[ActivityData]("user_activity", 10)

func main() {
    go UserActivity.Start(ctx)
    // ...
}
```

### 6.4 Signal Grouping

```go
// SignalGroup manages related signals together
type SignalGroup struct {
    signals []interface{}
}

// DisableGroup temporarily disables all signals in group
func (g *SignalGroup) Disable()

// EnableGroup re-enables signals
func (g *SignalGroup) Enable()

// Usage for testing
func TestWithoutSignals(t *testing.T) {
    signals.ModelSignals.Disable()
    defer signals.ModelSignals.Enable()
    
    // Tests run without triggering model signals
}
```

---

## 7. Error Handling

### 7.1 Handler Errors

```go
// Emit collects all handler errors
errors := signal.Emit(sender, data)
for _, err := range errors {
    log.Printf("Signal handler error: %v", err)
}
```

### 7.2 Error Strategies

```go
// Configure signal error behavior
type SignalOptions struct {
    // StopOnError halts emission on first error
    StopOnError bool
    
    // ErrorHandler custom error processing
    ErrorHandler func(handlerID string, err error)
    
    // RecoverPanic catches panics in handlers
    RecoverPanic bool
}

signal := signals.NewSignalWithOptions[T]("name", SignalOptions{
    RecoverPanic: true,
    ErrorHandler: func(id string, err error) {
        sentry.CaptureException(err)
    },
})
```

---

## 8. Testing Support

### 8.1 Signal Mocking

```go
import "github.com/yourusername/django-drf-go/core/signals/testing"

func TestUserRegistration(t *testing.T) {
    // Capture signal emissions
    recorder := testing.RecordSignal(signals.UserCreated)
    defer recorder.Stop()
    
    // Perform action
    user := createUser()
    
    // Assert signal was emitted
    assert.Equal(t, 1, recorder.Count())
    assert.Equal(t, user.ID, recorder.Emissions()[0].User.ID)
}
```

### 8.2 Disabling Signals

```go
func TestWithoutSignals(t *testing.T) {
    // Disable specific signal
    signals.UserCreated.Disable()
    defer signals.UserCreated.Enable()
    
    // Or globally
    signals.DisableAll()
    defer signals.EnableAll()
    
    // Run tests without signal side effects
}
```

---

## 9. Performance Considerations

### 9.1 Sync vs Async

| Mode | Use Case | Trade-off |
|------|----------|-----------|
| Sync (default) | Critical operations, validation | Blocks until complete |
| Async | Logging, analytics, notifications | Fire-and-forget |

### 9.2 Handler Timeout

```go
// Prevent slow handlers from blocking
errs := signal.EmitWithTimeout(sender, data, 5*time.Second)
```

### 9.3 Performance Targets

| Operation | Target |
|-----------|--------|
| Emit (no handlers) | < 50ns |
| Emit (1 handler) | < 500ns |
| Connect/Disconnect | < 1μs |

---

## 10. Thread Safety

- All signal operations are thread-safe
- Handlers can be connected/disconnected during emission
- Use `sync.RWMutex` for handler list operations
- Context propagation for cancellation

---

## 11. Implementation Notes

### 11.1 Dependencies

| Package | Purpose |
|---------|---------|
| `sync` | Mutex for handler management |
| `reflect` | Sender type filtering |
| `context` | Async cancellation |

### 11.2 Internal Structure

```go
type Signal[T any] struct {
    name     string
    handlers map[string]handlerEntry[T]
    order    []string  // Maintains insertion order
    disabled bool
    mu       sync.RWMutex
}

type handlerEntry[T any] struct {
    handler  Handler[T]
    priority int
    sender   reflect.Type
}
```

---

## 12. Related PRDs

- [01_PRD_Core_Apps_System.md](./01_PRD_Core_Apps_System.md) - Signal connection in Ready()
- [03_PRD_Core_Middleware.md](./03_PRD_Core_Middleware.md) - Request signals
- [07_PRD_ORM_Models.md](./07_PRD_ORM_Models.md) - Model signals
- [24_PRD_Contrib_Auth.md](./24_PRD_Contrib_Auth.md) - Auth signals
