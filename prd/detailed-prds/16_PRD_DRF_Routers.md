# PRD: DRF Routers

> **Module:** `drf/routers`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **DRF Equivalent:** `rest_framework.routers`

---

## 1. Overview

### 1.1 Purpose

Routers automatically generate URL patterns for ViewSets:

- **Automatic URL Generation**: RESTful endpoints from ViewSets
- **Nested Routers**: Parent-child resource relationships
- **Custom Actions**: Support for `@action` endpoints
- **Browsable API**: Endpoint discovery

---

## 2. Core Concepts

### 2.1 DefaultRouter

```go
// DefaultRouter provides standard REST URLs
router := drf.NewDefaultRouter()

// Register ViewSets
router.Register("users", &UserViewSet{})
router.Register("posts", &PostViewSet{})

// Generated URLs:
// GET/POST       /users/
// GET/PUT/DELETE /users/{id}/
// GET/POST       /posts/
// GET/PUT/DELETE /posts/{id}/
```

### 2.2 Generated URL Patterns

| Method | URL | Action |
|--------|-----|--------|
| GET | `/users/` | List |
| POST | `/users/` | Create |
| GET | `/users/{id}/` | Retrieve |
| PUT | `/users/{id}/` | Update |
| PATCH | `/users/{id}/` | Partial Update |
| DELETE | `/users/{id}/` | Destroy |
| GET/POST | `/users/{id}/action/` | Custom actions |

---

## 3. API Specification

### 3.1 Router Registration

```go
// Register a ViewSet
func (r *Router) Register(prefix string, viewset ViewSet, options ...RouteOption)

// Options
router.Register("users", &UserViewSet{}, 
    drf.WithBasename("user"),       // Override URL names
    drf.WithTrailingSlash(false),   // No trailing slash
)
```

### 3.2 Including in URL Config

```go
func SetupURLs() *urls.Router {
    apiRouter := drf.NewDefaultRouter()
    apiRouter.Register("users", &UserViewSet{})
    apiRouter.Register("posts", &PostViewSet{})
    
    main := urls.NewRouter()
    main.Include("/api/v1", apiRouter.URLs(), "api")
    
    return main
}
```

### 3.3 Custom Actions URLs

```go
// ViewSet with custom actions
type UserViewSet struct {
    drf.ModelViewSet[User]
}

func (v *UserViewSet) Actions() []drf.Action {
    return []drf.Action{
        {Name: "activate", Detail: true, Methods: []string{"POST"}},
        {Name: "recent", Detail: false, Methods: []string{"GET"}},
    }
}

// Generated:
// POST /users/{id}/activate/
// GET  /users/recent/
```

---

## 4. Nested Routers

```go
// Parent router
router := drf.NewDefaultRouter()
router.Register("users", &UserViewSet{})

// Nested router
postsRouter := drf.NewNestedRouter(router, "users")
postsRouter.Register("posts", &PostViewSet{})

// Generated:
// /users/{user_id}/posts/
// /users/{user_id}/posts/{id}/
```

### 4.1 Accessing Parent in ViewSet

```go
type PostViewSet struct {
    drf.ModelViewSet[Post]
}

func (v *PostViewSet) GetQuerySet(c *drf.Context) *orm.QuerySet[Post] {
    userID := c.Params["user_id"].(uint64)
    return orm.Objects[Post]().Filter(orm.Q{"user_id": userID})
}
```

---

## 5. DRF Comparison

**DRF:**
```python
from rest_framework.routers import DefaultRouter
router = DefaultRouter()
router.register('users', UserViewSet)

urlpatterns = [
    path('api/', include(router.urls)),
]
```

**Django-DRF-Go:**
```go
router := drf.NewDefaultRouter()
router.Register("users", &UserViewSet{})

main := urls.NewRouter()
main.Include("/api", router.URLs(), "api")
```

---

## 6. Related PRDs

- [04_PRD_Core_URL_Routing.md](./04_PRD_Core_URL_Routing.md) - URL patterns
- [15_PRD_DRF_Views_ViewSets.md](./15_PRD_DRF_Views_ViewSets.md) - ViewSets
