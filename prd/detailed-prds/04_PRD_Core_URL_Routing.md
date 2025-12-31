# PRD: Core URL Routing System

> **Module:** `core/urls`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **Django Equivalent:** `django.urls`

---

## 1. Overview

### 1.1 Purpose

The URL Routing System maps incoming HTTP requests to handlers:

- **Pattern Matching**: Path parameters, regex, and type converters
- **Namespacing**: Organize URLs by app with prefixes
- **Reverse URL Resolution**: Generate URLs from handler names
- **Nested Routes**: Include patterns from apps
- **HTTP Methods**: Method-based routing

### 1.2 Design Philosophy

| Django Approach | Go-Idiomatic Approach |
|-----------------|----------------------|
| `path()` and `re_path()` | `Path()` and `Pattern()` |
| String patterns with `<name:type>` | String patterns with `{name:type}` |
| `include()` for sub-patterns | `Include()` for sub-routers |
| Regex via `re_path` | Optional regex support |
| View functions or classes | `http.Handler` interface |

---

## 2. Core Concepts

### 2.1 URL Pattern

```go
// URLPattern defines a route
type URLPattern struct {
    // Pattern is the URL path with parameter placeholders
    // e.g., "/users/{id:int}", "/articles/{slug:slug}"
    Pattern string
    
    // Handler processes matching requests
    Handler http.Handler
    
    // Name for reverse URL resolution
    Name string
    
    // Methods allowed (empty = all methods)
    Methods []string
    
    // Middleware specific to this route
    Middleware []Middleware
}
```

### 2.2 Path Converters

```go
// Converter defines how to match and parse URL parameters
type Converter interface {
    // Regex returns the pattern to match
    Regex() string
    
    // ToGo converts string to Go type
    ToGo(value string) (interface{}, error)
    
    // ToURL converts Go type to URL string
    ToURL(value interface{}) (string, error)
}

// Built-in converters
var Converters = map[string]Converter{
    "int":    IntConverter{},      // Matches integers
    "str":    StringConverter{},   // Matches any non-slash string
    "slug":   SlugConverter{},     // Matches slug format
    "uuid":   UUIDConverter{},     // Matches UUID format
    "path":   PathConverter{},     // Matches any string including slashes
}
```

### 2.3 Router

```go
// Router manages URL patterns
type Router struct {
    patterns  []*URLPattern
    prefix    string
    namespace string
    
    // notFound for 404 handling
    notFound http.Handler
    
    // methodNotAllowed for 405 handling
    methodNotAllowed http.Handler
}

// NewRouter creates a new router
func NewRouter() *Router
```

---

## 3. API Specification

### 3.1 Defining Routes

#### Basic Routes

```go
import "github.com/yourusername/django-drf-go/core/urls"

func SetupURLs() *urls.Router {
    router := urls.NewRouter()
    
    // Simple path
    router.Path("/", HomeHandler, "home")
    
    // Path with parameters
    router.Path("/users/{id:int}", UserDetailHandler, "user-detail")
    router.Path("/articles/{slug:slug}", ArticleDetailHandler, "article-detail")
    
    // Method-specific routes
    router.Get("/api/items", ListItemsHandler, "item-list")
    router.Post("/api/items", CreateItemHandler, "item-create")
    router.Put("/api/items/{id:int}", UpdateItemHandler, "item-update")
    router.Delete("/api/items/{id:int}", DeleteItemHandler, "item-delete")
    
    return router
}
```

#### Pattern Methods

```go
// Path registers a route matching all methods
func (r *Router) Path(pattern string, handler http.Handler, name string) *Router

// Get/Post/Put/Patch/Delete for specific methods
func (r *Router) Get(pattern string, handler http.Handler, name string) *Router
func (r *Router) Post(pattern string, handler http.Handler, name string) *Router
func (r *Router) Put(pattern string, handler http.Handler, name string) *Router
func (r *Router) Patch(pattern string, handler http.Handler, name string) *Router
func (r *Router) Delete(pattern string, handler http.Handler, name string) *Router

// Methods for multiple specific methods
func (r *Router) Methods(methods []string, pattern string, handler http.Handler, name string) *Router

// Pattern for custom URLPattern configuration
func (r *Router) Pattern(pattern URLPattern) *Router
```

### 3.2 Including Sub-Routers

```go
// Include adds all patterns from another router under a prefix
func (r *Router) Include(prefix string, router *Router, namespace string) *Router

// Example:
func SetupURLs() *urls.Router {
    main := urls.NewRouter()
    
    // Include app routers
    main.Include("/api/users", users.GetURLs(), "users")
    main.Include("/api/orders", orders.GetURLs(), "orders")
    main.Include("/admin", admin.GetURLs(), "admin")
    
    return main
}

// In users/urls.go
func GetURLs() *urls.Router {
    router := urls.NewRouter()
    router.Get("", ListUsersHandler, "list")
    router.Post("", CreateUserHandler, "create")
    router.Get("/{id:int}", DetailUserHandler, "detail")
    router.Put("/{id:int}", UpdateUserHandler, "update")
    router.Delete("/{id:int}", DeleteUserHandler, "delete")
    return router
}
```

### 3.3 Reverse URL Resolution

```go
// Reverse generates a URL from a name and parameters
func (r *Router) Reverse(name string, args ...interface{}) (string, error)

// ReverseWithQuery includes query parameters
func (r *Router) ReverseWithQuery(name string, query url.Values, args ...interface{}) (string, error)

// MustReverse panics on error (for templates)
func (r *Router) MustReverse(name string, args ...interface{}) string

// Examples:
url, _ := router.Reverse("user-detail", 42)           // "/users/42"
url, _ := router.Reverse("users:detail", 42)          // "/api/users/42" (namespaced)
url, _ := router.Reverse("article-detail", "hello-world")  // "/articles/hello-world"
```

### 3.4 Accessing URL Parameters

```go
// URLParams returns parsed parameters from request context
func URLParams(r *http.Request) map[string]interface{}

// URLParam returns a single parameter value
func URLParam[T any](r *http.Request, name string) (T, error)

// Example handler:
func UserDetailHandler(w http.ResponseWriter, r *http.Request) {
    userID, err := urls.URLParam[int](r, "id")
    if err != nil {
        http.Error(w, "Invalid user ID", 400)
        return
    }
    
    // Or get all params
    params := urls.URLParams(r)
    userID := params["id"].(int)
}
```

---

## 4. Path Converters

### 4.1 Built-in Converters

| Converter | Regex | Description | Example |
|-----------|-------|-------------|---------|
| `str` | `[^/]+` | Any string without slashes | `hello`, `my-name` |
| `int` | `[0-9]+` | Positive integers | `42`, `1234` |
| `slug` | `[-a-zA-Z0-9_]+` | Slug-style strings | `hello-world` |
| `uuid` | UUID pattern | UUID v4 | `550e8400-e29b-...` |
| `path` | `.+` | Any string including slashes | `path/to/file` |

### 4.2 Custom Converters

```go
// DateConverter for date parameters
type DateConverter struct{}

func (c DateConverter) Regex() string {
    return `\d{4}-\d{2}-\d{2}`  // YYYY-MM-DD
}

func (c DateConverter) ToGo(value string) (interface{}, error) {
    return time.Parse("2006-01-02", value)
}

func (c DateConverter) ToURL(value interface{}) (string, error) {
    if t, ok := value.(time.Time); ok {
        return t.Format("2006-01-02"), nil
    }
    return "", errors.New("expected time.Time")
}

// Register custom converter
func init() {
    urls.RegisterConverter("date", DateConverter{})
}

// Use in patterns
router.Path("/events/{date:date}", EventHandler, "event-by-date")
```

---

## 5. Django Comparison

### 5.1 Basic URL Configuration

**Django (urls.py):**
```python
from django.urls import path
from . import views

urlpatterns = [
    path('', views.home, name='home'),
    path('users/<int:id>/', views.user_detail, name='user-detail'),
    path('articles/<slug:slug>/', views.article_detail, name='article-detail'),
]
```

**Django-DRF-Go:**
```go
// urls.go
package main

import "github.com/yourusername/django-drf-go/core/urls"

func SetupURLs() *urls.Router {
    router := urls.NewRouter()
    
    router.Path("/", HomeHandler, "home")
    router.Path("/users/{id:int}", UserDetailHandler, "user-detail")
    router.Path("/articles/{slug:slug}", ArticleDetailHandler, "article-detail")
    
    return router
}
```

### 5.2 Including URLs

**Django:**
```python
# main/urls.py
from django.urls import path, include

urlpatterns = [
    path('api/users/', include('users.urls', namespace='users')),
    path('api/orders/', include('orders.urls', namespace='orders')),
    path('admin/', admin.site.urls),
]
```

**Django-DRF-Go:**
```go
func SetupURLs() *urls.Router {
    router := urls.NewRouter()
    
    router.Include("/api/users", users.GetURLs(), "users")
    router.Include("/api/orders", orders.GetURLs(), "orders")
    router.Include("/admin", admin.GetURLs(), "admin")
    
    return router
}
```

### 5.3 Reverse URL Resolution

**Django:**
```python
from django.urls import reverse

url = reverse('user-detail', args=[42])
url = reverse('users:detail', args=[42])
url = reverse('article-detail', kwargs={'slug': 'hello-world'})
```

**Django-DRF-Go:**
```go
import "github.com/yourusername/django-drf-go/core/urls"

url, _ := urls.Reverse("user-detail", 42)
url, _ := urls.Reverse("users:detail", 42)
url, _ := urls.Reverse("article-detail", "hello-world")
```

### 5.4 Path Converters

**Django:**
```python
path('events/<date:date>/', views.event, name='event')

# Custom converter
class DateConverter:
    regex = r'\d{4}-\d{2}-\d{2}'
    
    def to_python(self, value):
        return datetime.strptime(value, '%Y-%m-%d').date()
    
    def to_url(self, value):
        return value.strftime('%Y-%m-%d')

register_converter(DateConverter, 'date')
```

**Django-DRF-Go:**
```go
router.Path("/events/{date:date}", EventHandler, "event")

// Custom converter (see section 4.2)
```

---

## 6. Advanced Features

### 6.1 Regex Patterns

```go
// RegexPath for complex patterns
router.RegexPath(`/archive/(?P<year>\d{4})/(?P<month>\d{2})`, ArchiveHandler, "archive")

// Access named groups
params := urls.URLParams(r)
year := params["year"].(string)
month := params["month"].(string)
```

### 6.2 Route Groups

```go
// Group applies middleware to multiple routes
router.Group(authMiddleware, func(r *urls.Router) {
    r.Get("/profile", ProfileHandler, "profile")
    r.Put("/profile", UpdateProfileHandler, "profile-update")
    r.Get("/settings", SettingsHandler, "settings")
})
```

### 6.3 Route Constraints

```go
// Add constraints to routes
router.Path("/users/{id:int}", UserHandler, "user").
    Where("id", func(v int) bool { return v > 0 })
    
// Host-based routing
router.Host("api.example.com", apiRouter)
router.Host("{subdomain}.example.com", subdomainRouter)
```

### 6.4 Static and Media Files

```go
// Static file serving
router.Static("/static/", "./static")

// Media file serving (uploads)
router.Static("/media/", settings.Get().MediaRoot)
```

---

## 7. Router as http.Handler

```go
// Router implements http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request)

// Use with http.Server
func main() {
    router := SetupURLs()
    
    // Apply global middleware
    handler := middleware.Chain(
        logging.Middleware,
        recovery.Middleware,
    ).Then(router)
    
    http.ListenAndServe(":8000", handler)
}
```

---

## 8. Error Handling

### 8.1 Not Found (404)

```go
// Custom 404 handler
router.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(404)
    json.NewEncoder(w).Encode(map[string]string{
        "error": "Not found",
        "path":  r.URL.Path,
    })
}))
```

### 8.2 Method Not Allowed (405)

```go
// Custom 405 handler
router.MethodNotAllowed(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Allow", "GET, POST")  // Set by router automatically
    w.WriteHeader(405)
    json.NewEncoder(w).Encode(map[string]string{
        "error": "Method not allowed",
    })
}))
```

---

## 9. URL Generation Helpers

### 9.1 Template Integration

```go
// URL function for templates
func TemplateURL(name string, args ...interface{}) string {
    return urls.MustReverse(name, args...)
}

// Register with templates
tmpl.Funcs(template.FuncMap{
    "url": TemplateURL,
})
```

```html
<!-- In templates -->
<a href="{{ url "user-detail" .User.ID }}">View Profile</a>
<form action="{{ url "users:update" .User.ID }}" method="post">
```

### 9.2 Redirect Helpers

```go
// RedirectToURL redirects to a named URL
func RedirectToURL(w http.ResponseWriter, r *http.Request, name string, args ...interface{}) {
    url, _ := urls.Reverse(name, args...)
    http.Redirect(w, r, url, http.StatusFound)
}

// Usage in handler
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    user := createUser(r)
    urls.RedirectToURL(w, r, "user-detail", user.ID)
}
```

---

## 10. Thread Safety

- Router is **immutable after setup** - no locking needed during requests
- Pattern compilation happens once at startup
- URL parameters are stored in request context (per-request isolation)

---

## 11. Performance

### 11.1 Pattern Matching

- Patterns compiled to efficient matchers at startup
- O(n) pattern matching where n = number of routes
- Consider radix tree for large route tables

### 11.2 Benchmarks

| Operation | Target |
|-----------|--------|
| Simple path match | < 100ns |
| Parameter extraction | < 200ns |
| Reverse URL resolution | < 150ns |

---

## 12. Related PRDs

- [03_PRD_Core_Middleware.md](./03_PRD_Core_Middleware.md) - Middleware integration
- [15_PRD_DRF_Views_ViewSets.md](./15_PRD_DRF_Views_ViewSets.md) - ViewSet routing
- [16_PRD_DRF_Routers.md](./16_PRD_DRF_Routers.md) - Automatic API routing
- [21_PRD_Admin_Dashboard.md](./21_PRD_Admin_Dashboard.md) - Admin URL setup
