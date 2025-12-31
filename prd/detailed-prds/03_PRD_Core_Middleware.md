# PRD: Core Middleware System

> **Module:** `core/middleware`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **Django Equivalent:** `django.middleware`

---

## 1. Overview

### 1.1 Purpose

The Middleware System provides a request/response processing pipeline:

- **Request Processing**: Transform, validate, or reject incoming requests
- **Response Processing**: Modify outgoing responses (headers, content, etc.)
- **Cross-Cutting Concerns**: Authentication, CORS, logging, compression
- **Error Handling**: Catch and process exceptions uniformly
- **Lifecycle Hooks**: Execute code before/after views

### 1.2 Design Philosophy

| Django Approach | Go-Idiomatic Approach |
|-----------------|----------------------|
| Class-based middleware with `__call__` | `http.Handler` wrapper pattern |
| `process_request`, `process_response` | Single function with next handler |
| Class instantiation per-request | Functional middleware, stateless |
| Order via list in settings | Chain builder pattern |

---

## 2. Core Concepts

### 2.1 Middleware Function Signature

```go
// Middleware wraps an http.Handler with additional logic
type Middleware func(http.Handler) http.Handler

// MiddlewareFunc for inline middleware definitions
type MiddlewareFunc func(w http.ResponseWriter, r *http.Request, next http.Handler)
```

### 2.2 Middleware Chain

```go
// Chain composes multiple middleware into a single handler
type Chain struct {
    middlewares []Middleware
}

// New creates a middleware chain
func New(middlewares ...Middleware) *Chain

// Then wraps the final handler with all middleware
func (c *Chain) Then(handler http.Handler) http.Handler

// ThenFunc wraps an http.HandlerFunc
func (c *Chain) ThenFunc(fn http.HandlerFunc) http.Handler

// Append adds middleware to the chain
func (c *Chain) Append(middlewares ...Middleware) *Chain
```

### 2.3 Request Context

```go
// Context keys for middleware-stored values
type contextKey string

const (
    RequestIDKey   contextKey = "request_id"
    UserKey        contextKey = "user"
    SessionKey     contextKey = "session"
    StartTimeKey   contextKey = "start_time"
)

// GetRequestID retrieves request ID from context
func GetRequestID(r *http.Request) string

// GetUser retrieves authenticated user from context
func GetUser(r *http.Request) *auth.User

// GetSession retrieves session from context
func GetSession(r *http.Request) *sessions.Session
```

---

## 3. Built-in Middleware

### 3.1 Security Middleware

```go
// SecurityMiddleware adds security headers
func SecurityMiddleware(next http.Handler) http.Handler

// Configurable via settings:
type SecurityConfig struct {
    // XSSProtection enables X-XSS-Protection header
    XSSProtection bool `json:"xss_protection" default:"true"`
    
    // ContentTypeNosniff prevents content type sniffing
    ContentTypeNosniff bool `json:"content_type_nosniff" default:"true"`
    
    // XFrameOptions prevents clickjacking
    XFrameOptions string `json:"x_frame_options" default:"DENY"`
    
    // HSTSSeconds enables HSTS (0 to disable)
    HSTSSeconds int `json:"hsts_seconds" default:"31536000"`
    
    // HSTSIncludeSubdomains includes subdomains in HSTS
    HSTSIncludeSubdomains bool `json:"hsts_include_subdomains" default:"true"`
    
    // ContentSecurityPolicy header value
    ContentSecurityPolicy string `json:"content_security_policy"`
    
    // ReferrerPolicy header value
    ReferrerPolicy string `json:"referrer_policy" default:"same-origin"`
}
```

### 3.2 CORS Middleware

```go
// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware(config CORSConfig) Middleware

type CORSConfig struct {
    // AllowedOrigins list (use "*" for all)
    AllowedOrigins []string `json:"allowed_origins"`
    
    // AllowedMethods for preflight
    AllowedMethods []string `json:"allowed_methods" default:"GET,POST,PUT,DELETE,OPTIONS"`
    
    // AllowedHeaders for preflight
    AllowedHeaders []string `json:"allowed_headers"`
    
    // ExposedHeaders visible to client
    ExposedHeaders []string `json:"exposed_headers"`
    
    // AllowCredentials includes cookies
    AllowCredentials bool `json:"allow_credentials"`
    
    // MaxAge for preflight cache (seconds)
    MaxAge int `json:"max_age" default:"86400"`
}
```

### 3.3 Common Middleware

```go
// CommonMiddleware handles common HTTP concerns
func CommonMiddleware(next http.Handler) http.Handler

// Provides:
// - Trailing slash normalization
// - WWW redirect
// - ETag handling
// - Content-Length validation
```

### 3.4 Session Middleware

```go
// SessionMiddleware manages user sessions
func SessionMiddleware(config SessionConfig) Middleware

type SessionConfig struct {
    // Store backend for sessions
    Store sessions.Store
    
    // CookieName for session ID
    CookieName string `default:"sessionid"`
    
    // MaxAge in seconds (0 = browser session)
    MaxAge int `default:"1209600"` // 2 weeks
    
    // Secure requires HTTPS
    Secure bool
    
    // HttpOnly prevents JavaScript access
    HttpOnly bool `default:"true"`
    
    // SameSite cookie policy
    SameSite http.SameSite `default:"Lax"`
}
```

### 3.5 CSRF Middleware

```go
// CSRFMiddleware protects against CSRF attacks
func CSRFMiddleware(config CSRFConfig) Middleware

type CSRFConfig struct {
    // TokenLength for generated tokens
    TokenLength int `default:"32"`
    
    // CookieName for CSRF token
    CookieName string `default:"csrftoken"`
    
    // HeaderName to check for token
    HeaderName string `default:"X-CSRFToken"`
    
    // SafeMethods don't require CSRF
    SafeMethods []string `default:"GET,HEAD,OPTIONS,TRACE"`
    
    // TrustedOrigins for cross-origin requests
    TrustedOrigins []string
}

// GetCSRFToken retrieves token for templates
func GetCSRFToken(r *http.Request) string
```

### 3.6 Authentication Middleware

```go
// AuthenticationMiddleware populates request.user
func AuthenticationMiddleware(backends ...AuthBackend) Middleware

// AuthBackend interface for pluggable authentication
type AuthBackend interface {
    Authenticate(r *http.Request) (*auth.User, error)
}
```

### 3.7 Request ID Middleware

```go
// RequestIDMiddleware adds unique request ID
func RequestIDMiddleware(next http.Handler) http.Handler

// Uses X-Request-ID header or generates UUID
```

### 3.8 Logging Middleware

```go
// LoggingMiddleware logs requests and responses
func LoggingMiddleware(logger *slog.Logger) Middleware

// Logs: method, path, status, duration, request ID
```

### 3.9 Recovery Middleware

```go
// RecoveryMiddleware catches panics
func RecoveryMiddleware(config RecoveryConfig) Middleware

type RecoveryConfig struct {
    // Logger for panic messages
    Logger *slog.Logger
    
    // StackTrace includes stack in logs
    StackTrace bool
    
    // Handler for custom error response
    Handler func(w http.ResponseWriter, r *http.Request, err interface{})
}
```

### 3.10 Compression Middleware

```go
// CompressionMiddleware compresses responses
func CompressionMiddleware(config CompressionConfig) Middleware

type CompressionConfig struct {
    // Level of compression (1-9)
    Level int `default:"5"`
    
    // MinSize to trigger compression
    MinSize int `default:"1024"`
    
    // Types to compress
    Types []string `default:"text/html,application/json,text/css,text/javascript"`
}
```

---

## 4. API Specification

### 4.1 Creating Middleware

#### Standard Pattern

```go
// Pattern 1: Function returning Middleware
func MyMiddleware(config MyConfig) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Pre-processing
            log.Println("Before:", r.URL.Path)
            
            // Call next handler
            next.ServeHTTP(w, r)
            
            // Post-processing
            log.Println("After:", r.URL.Path)
        })
    }
}

// Pattern 2: Simple middleware function
func TimingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("Request took %v", time.Since(start))
    })
}
```

#### Response Capture Pattern

```go
// For modifying responses, use ResponseWriter wrapper
func TransformMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Wrap response writer to capture
        rec := middleware.NewResponseRecorder(w)
        
        next.ServeHTTP(rec, r)
        
        // Modify captured response
        body := rec.Body()
        modified := transform(body)
        
        // Write modified response
        w.Header().Set("Content-Length", strconv.Itoa(len(modified)))
        w.WriteHeader(rec.StatusCode())
        w.Write(modified)
    })
}
```

### 4.2 Using Middleware Chain

```go
// Build chain from settings
func BuildMiddlewareChain() *Chain {
    chain := middleware.New(
        middleware.RecoveryMiddleware(RecoveryConfig{}),
        middleware.RequestIDMiddleware,
        middleware.LoggingMiddleware(logger),
        middleware.SecurityMiddleware,
        middleware.CORSMiddleware(corsConfig),
        middleware.SessionMiddleware(sessionConfig),
        middleware.AuthenticationMiddleware(backends...),
        middleware.CSRFMiddleware(csrfConfig),
    )
    return chain
}

// Apply to router
handler := chain.Then(router)
http.ListenAndServe(":8000", handler)
```

### 4.3 Per-Route Middleware

```go
// Apply middleware to specific routes
adminMiddleware := middleware.New(
    middleware.IsAuthenticatedMiddleware,
    middleware.IsStaffMiddleware,
)

router.Handle("/admin/", adminMiddleware.Then(adminRouter))
router.Handle("/api/", apiChain.Then(apiRouter))
```

---

## 5. Django Comparison

### 5.1 Middleware Definition

**Django:**
```python
# middleware.py
class TimingMiddleware:
    def __init__(self, get_response):
        self.get_response = get_response

    def __call__(self, request):
        start = time.time()
        response = self.get_response(request)
        duration = time.time() - start
        response['X-Request-Duration'] = str(duration)
        return response
```

**Django-DRF-Go:**
```go
// middleware.go
func TimingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Wrap to capture status
        rec := middleware.NewResponseRecorder(w)
        next.ServeHTTP(rec, r)
        
        duration := time.Since(start)
        w.Header().Set("X-Request-Duration", duration.String())
        
        rec.Flush(w)
    })
}
```

### 5.2 Configuration

**Django (settings.py):**
```python
MIDDLEWARE = [
    'django.middleware.security.SecurityMiddleware',
    'django.contrib.sessions.middleware.SessionMiddleware',
    'django.middleware.common.CommonMiddleware',
    'django.middleware.csrf.CsrfViewMiddleware',
    'django.contrib.auth.middleware.AuthenticationMiddleware',
]
```

**Django-DRF-Go (settings.json):**
```json
{
  "middleware": [
    "security.SecurityMiddleware",
    "sessions.SessionMiddleware",
    "common.CommonMiddleware",
    "csrf.CSRFMiddleware",
    "auth.AuthenticationMiddleware"
  ]
}
```

### 5.3 Request/Response Hooks

**Django:**
```python
class MyMiddleware:
    def process_request(self, request):
        # Before view
        pass
    
    def process_response(self, request, response):
        # After view
        return response
    
    def process_exception(self, request, exception):
        # On exception
        pass
```

**Django-DRF-Go:**
```go
func MyMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // process_request equivalent
        beforeView(r)
        
        // Wrap for exception handling
        defer func() {
            if err := recover(); err != nil {
                // process_exception equivalent
                handleException(w, r, err)
            }
        }()
        
        rec := middleware.NewResponseRecorder(w)
        next.ServeHTTP(rec, r)
        
        // process_response equivalent
        afterView(r, rec)
        rec.Flush(w)
    })
}
```

---

## 6. Response Utilities

### 6.1 Response Recorder

```go
// ResponseRecorder captures response for modification
type ResponseRecorder struct {
    http.ResponseWriter
    statusCode int
    body       *bytes.Buffer
    headers    http.Header
    written    bool
}

func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder

func (r *ResponseRecorder) StatusCode() int
func (r *ResponseRecorder) Body() []byte
func (r *ResponseRecorder) Flush(w http.ResponseWriter)
```

### 6.2 Response Helpers

```go
// SendJSON writes JSON response
func SendJSON(w http.ResponseWriter, status int, data interface{})

// SendError writes error response
func SendError(w http.ResponseWriter, status int, message string)

// Redirect sends redirect response
func Redirect(w http.ResponseWriter, r *http.Request, url string, code int)
```

---

## 7. Middleware Ordering

### 7.1 Recommended Order

```
┌─────────────────────────────────────────────────────────────┐
│ 1. RecoveryMiddleware     - Catch panics first              │
│ 2. RequestIDMiddleware    - Tag all logs with ID            │
│ 3. LoggingMiddleware      - Log all requests                │
│ 4. SecurityMiddleware     - Add security headers            │
│ 5. CORSMiddleware         - Handle preflight                │
│ 6. CompressionMiddleware  - Compress responses              │
│ 7. SessionMiddleware      - Load session                    │
│ 8. AuthenticationMiddleware - Set request.user              │
│ 9. CSRFMiddleware         - Verify CSRF tokens              │
│ 10. CommonMiddleware      - URL normalization               │
└─────────────────────────────────────────────────────────────┘
                              ↓
                        Router/Views
```

### 7.2 Order Matters

| Middleware | Must Be Before | Reason |
|------------|---------------|--------|
| Recovery | All others | Catch panics anywhere |
| RequestID | Logging | Include ID in logs |
| CORS | Authentication | Preflight doesn't include auth |
| Session | Authentication | Auth may need session |
| Authentication | CSRF | CSRF checks authenticated users |

---

## 8. Error Handling

### 8.1 Exception Middleware

```go
// ExceptionMiddleware converts panics to HTTP errors
func ExceptionMiddleware(config ExceptionConfig) Middleware

type ExceptionConfig struct {
    // Debug shows stack traces
    Debug bool
    
    // Logger for errors
    Logger *slog.Logger
    
    // Handlers for specific error types
    Handlers map[reflect.Type]ExceptionHandler
}

type ExceptionHandler func(w http.ResponseWriter, r *http.Request, err error)
```

### 8.2 HTTP Exception Types

```go
// HTTP exceptions for common errors
type HTTPException struct {
    StatusCode int
    Message    string
    Detail     interface{}
}

func NewBadRequest(message string) *HTTPException
func NewUnauthorized(message string) *HTTPException
func NewForbidden(message string) *HTTPException
func NewNotFound(message string) *HTTPException
func NewMethodNotAllowed(message string) *HTTPException
```

---

## 9. Testing Support

### 9.1 Test Helpers

```go
import "github.com/yourusername/django-drf-go/core/middleware/testing"

func TestMyMiddleware(t *testing.T) {
    // Create test chain
    chain := middleware.New(MyMiddleware)
    
    handler := chain.ThenFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })
    
    // Create test request
    req := httptest.NewRequest("GET", "/test", nil)
    rec := httptest.NewRecorder()
    
    // Execute
    handler.ServeHTTP(rec, req)
    
    // Assert
    assert.Equal(t, 200, rec.Code)
}
```

### 9.2 Mock Middleware

```go
// MockMiddleware tracks calls for testing
type MockMiddleware struct {
    Called    bool
    Request   *http.Request
    CallNext  bool  // Whether to call next handler
}

func (m *MockMiddleware) Handler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        m.Called = true
        m.Request = r
        if m.CallNext {
            next.ServeHTTP(w, r)
        }
    })
}
```

---

## 10. Implementation Notes

### 10.1 Performance

- Middleware allocation happens once at startup, not per-request
- Use sync.Pool for response recorders in high-traffic scenarios
- Avoid allocations in hot paths

### 10.2 Context Best Practices

```go
// Do: Use typed context keys
type contextKey string
const userKey contextKey = "user"

// Don't: Use string keys directly
// ctx = context.WithValue(ctx, "user", user)  // Bad!
```

### 10.3 Dependencies

| Package | Purpose |
|---------|---------|
| `net/http` | HTTP primitives |
| `context` | Request-scoped values |
| `compress/gzip` | Response compression |
| `crypto/rand` | Secure token generation |

---

## 11. Related PRDs

- [02_PRD_Core_Settings.md](./02_PRD_Core_Settings.md) - Middleware configuration
- [04_PRD_Core_URL_Routing.md](./04_PRD_Core_URL_Routing.md) - Router integration
- [17_PRD_DRF_Authentication.md](./17_PRD_DRF_Authentication.md) - Auth middleware
- [25_PRD_Contrib_Sessions.md](./25_PRD_Contrib_Sessions.md) - Session middleware
