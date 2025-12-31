# PRD: DRF Views and ViewSets

> **Module:** `drf/views`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **DRF Equivalent:** `rest_framework.views`, `rest_framework.viewsets`

---

## 1. Overview

### 1.1 Purpose

The Views System handles HTTP requests and responses:

- **APIView**: Class-based views for HTTP methods
- **ViewSet**: Group related views for a resource
- **Mixins**: Reusable CRUD operations
- **Actions**: Custom endpoints on ViewSets

### 1.2 Design Philosophy

| DRF Approach | Go-Idiomatic Approach |
|--------------|----------------------|
| `APIView` class | Struct implementing methods |
| `ViewSet` class | Struct with action methods |
| `@action` decorator | `Action{}` registration |
| `self.request` | Request passed as parameter |

---

## 2. Core Concepts

### 2.1 APIView Interface

```go
// View handles HTTP requests
type View interface {
    // HTTP methods (optional - implement only needed ones)
    Get(c *Context) Response
    Post(c *Context) Response
    Put(c *Context) Response
    Patch(c *Context) Response
    Delete(c *Context) Response
    
    // Configuration
    GetPermissions() []Permission
    GetAuthentication() []Authentication
    GetThrottles() []Throttle
}
```

### 2.2 Context

```go
// Context provides request context to views
type Context struct {
    Request        *http.Request
    ResponseWriter http.ResponseWriter
    User           *auth.User
    Auth           interface{}
    Params         map[string]interface{}  // URL params
    Query          url.Values              // Query params
    Data           map[string]interface{}  // Parsed body
}

// Methods
func (c *Context) GetObject(key string) interface{}
func (c *Context) SetObject(key string, value interface{})
```

### 2.3 Response

```go
// Response represents HTTP response
type Response struct {
    Status      int
    Data        interface{}
    Headers     map[string]string
    ContentType string
}

// Response helpers
func OK(data interface{}) Response
func Created(data interface{}) Response
func NoContent() Response
func BadRequest(errors interface{}) Response
func Unauthorized(message string) Response
func Forbidden(message string) Response
func NotFound(message string) Response
func MethodNotAllowed() Response
func InternalError(message string) Response
```

---

## 3. APIView

### 3.1 Basic APIView

```go
type UserListView struct {
    drf.APIView
}

func (v *UserListView) Get(c *drf.Context) drf.Response {
    users, err := orm.Objects[User]().All()
    if err != nil {
        return drf.InternalError(err.Error())
    }
    
    data, _ := drf.SerializeMany[UserSerializer](users)
    return drf.OK(data)
}

func (v *UserListView) Post(c *drf.Context) drf.Response {
    serializer := &CreateUserSerializer{}
    if err := serializer.Validate(c.Data); err != nil {
        return drf.BadRequest(serializer.Errors())
    }
    
    user, err := serializer.Create(serializer.ValidatedData())
    if err != nil {
        return drf.InternalError(err.Error())
    }
    
    data, _ := drf.Serialize[UserSerializer](user)
    return drf.Created(data)
}

// Configuration
func (v *UserListView) GetPermissions() []drf.Permission {
    return []drf.Permission{drf.IsAuthenticated()}
}
```

### 3.2 Detail View

```go
type UserDetailView struct {
    drf.APIView
}

func (v *UserDetailView) Get(c *drf.Context) drf.Response {
    userID := c.Params["id"].(uint64)
    
    user, err := orm.Objects[User]().Get(orm.Q{"id": userID})
    if err != nil {
        return drf.NotFound("User not found")
    }
    
    data, _ := drf.Serialize[UserSerializer](user)
    return drf.OK(data)
}

func (v *UserDetailView) Put(c *drf.Context) drf.Response {
    userID := c.Params["id"].(uint64)
    
    user, err := orm.Objects[User]().Get(orm.Q{"id": userID})
    if err != nil {
        return drf.NotFound("User not found")
    }
    
    serializer := &UpdateUserSerializer{}
    serializer.SetInstance(user)
    
    if err := serializer.Validate(c.Data); err != nil {
        return drf.BadRequest(serializer.Errors())
    }
    
    user, err = serializer.Update(user, serializer.ValidatedData())
    if err != nil {
        return drf.InternalError(err.Error())
    }
    
    data, _ := drf.Serialize[UserSerializer](user)
    return drf.OK(data)
}

func (v *UserDetailView) Delete(c *drf.Context) drf.Response {
    userID := c.Params["id"].(uint64)
    
    user, err := orm.Objects[User]().Get(orm.Q{"id": userID})
    if err != nil {
        return drf.NotFound("User not found")
    }
    
    if err := user.Delete(); err != nil {
        return drf.InternalError(err.Error())
    }
    
    return drf.NoContent()
}
```

---

## 4. ViewSet

### 4.1 Basic ViewSet

```go
// ViewSet groups actions for a resource
type UserViewSet struct {
    drf.ViewSet[User]
}

// List - GET /users/
func (v *UserViewSet) List(c *drf.Context) drf.Response {
    users := v.GetQuerySet().All()
    data := v.Serialize(users)
    return drf.OK(data)
}

// Create - POST /users/
func (v *UserViewSet) Create(c *drf.Context) drf.Response {
    serializer := v.GetSerializer(c.Data)
    if !serializer.IsValid() {
        return drf.BadRequest(serializer.Errors())
    }
    user := serializer.Save()
    return drf.Created(v.Serialize(user))
}

// Retrieve - GET /users/{id}/
func (v *UserViewSet) Retrieve(c *drf.Context) drf.Response {
    user := v.GetObject(c)
    return drf.OK(v.Serialize(user))
}

// Update - PUT /users/{id}/
func (v *UserViewSet) Update(c *drf.Context) drf.Response {
    user := v.GetObject(c)
    serializer := v.GetSerializer(c.Data, user)
    if !serializer.IsValid() {
        return drf.BadRequest(serializer.Errors())
    }
    user = serializer.Save()
    return drf.OK(v.Serialize(user))
}

// PartialUpdate - PATCH /users/{id}/
func (v *UserViewSet) PartialUpdate(c *drf.Context) drf.Response {
    // Same as Update but with partial=true
    return v.Update(c)
}

// Destroy - DELETE /users/{id}/
func (v *UserViewSet) Destroy(c *drf.Context) drf.Response {
    user := v.GetObject(c)
    user.Delete()
    return drf.NoContent()
}
```

### 4.2 ModelViewSet

```go
// ModelViewSet provides full CRUD automatically
type UserViewSet struct {
    drf.ModelViewSet[User]
}

func (v *UserViewSet) Meta() drf.ViewSetMeta {
    return drf.ViewSetMeta{
        QuerySet:        orm.Objects[User](),
        Serializer:      &UserSerializer{},
        Permissions:     []drf.Permission{drf.IsAuthenticated()},
        FilterBackends:  []drf.FilterBackend{drf.SearchFilter(), drf.OrderingFilter()},
        SearchFields:    []string{"username", "email"},
        OrderingFields:  []string{"created_at", "username"},
    }
}

// All CRUD operations are automatic!
// Override only if customization needed
```

### 4.3 ReadOnlyModelViewSet

```go
// ReadOnlyModelViewSet only provides List and Retrieve
type ArticleViewSet struct {
    drf.ReadOnlyModelViewSet[Article]
}

func (v *ArticleViewSet) Meta() drf.ViewSetMeta {
    return drf.ViewSetMeta{
        QuerySet:   orm.Objects[Article]().Filter(orm.Q{"is_published": true}),
        Serializer: &ArticleSerializer{},
    }
}
```

---

## 5. Mixins

### 5.1 Available Mixins

```go
// Mixins provide reusable behavior
type CreateMixin      // Create action
type ListMixin        // List action
type RetrieveMixin    // Retrieve action
type UpdateMixin      // Update action (full + partial)
type DestroyMixin     // Destroy action

// Combine for custom ViewSets
type ProductViewSet struct {
    drf.GenericViewSet[Product]
    drf.ListMixin[Product]
    drf.RetrieveMixin[Product]
    drf.CreateMixin[Product]
    // No Update or Destroy
}
```

### 5.2 Mixin Implementation

```go
// ListMixin implementation
type ListMixin[T orm.ModelInterface] struct{}

func (m *ListMixin[T]) List(c *drf.Context) drf.Response {
    v := c.View.(GenericViewSetInterface[T])
    
    queryset := v.GetQuerySet()
    queryset = v.FilterQuerySet(queryset, c)
    
    page := v.Paginate(queryset, c)
    if page != nil {
        data := v.SerializeMany(page.Results)
        return v.GetPaginatedResponse(data)
    }
    
    data := v.SerializeMany(queryset.All())
    return drf.OK(data)
}
```

---

## 6. Custom Actions

```go
type UserViewSet struct {
    drf.ModelViewSet[User]
}

// Actions returns custom actions
func (v *UserViewSet) Actions() []drf.Action {
    return []drf.Action{
        {
            Name:    "activate",
            Methods: []string{"POST"},
            Detail:  true,  // /users/{id}/activate/
            Handler: v.Activate,
        },
        {
            Name:    "deactivate",
            Methods: []string{"POST"},
            Detail:  true,  // /users/{id}/deactivate/
            Handler: v.Deactivate,
        },
        {
            Name:    "recent",
            Methods: []string{"GET"},
            Detail:  false,  // /users/recent/
            Handler: v.Recent,
        },
    }
}

func (v *UserViewSet) Activate(c *drf.Context) drf.Response {
    user := v.GetObject(c)
    user.IsActive = true
    user.Save()
    return drf.OK(v.Serialize(user))
}

func (v *UserViewSet) Deactivate(c *drf.Context) drf.Response {
    user := v.GetObject(c)
    user.IsActive = false
    user.Save()
    return drf.OK(v.Serialize(user))
}

func (v *UserViewSet) Recent(c *drf.Context) drf.Response {
    users := v.GetQuerySet().
        Filter(orm.Q{"created_at__gte": time.Now().AddDate(0, 0, -7)}).
        OrderBy("-created_at").
        Limit(10).
        All()
    return drf.OK(v.SerializeMany(users))
}
```

---

## 7. DRF Comparison

### 7.1 APIView

**DRF:**
```python
from rest_framework.views import APIView
from rest_framework.response import Response
from rest_framework import status

class UserListView(APIView):
    permission_classes = [IsAuthenticated]
    
    def get(self, request):
        users = User.objects.all()
        serializer = UserSerializer(users, many=True)
        return Response(serializer.data)
    
    def post(self, request):
        serializer = UserSerializer(data=request.data)
        if serializer.is_valid():
            serializer.save()
            return Response(serializer.data, status=status.HTTP_201_CREATED)
        return Response(serializer.errors, status=status.HTTP_400_BAD_REQUEST)
```

**Django-DRF-Go:**
```go
type UserListView struct {
    drf.APIView
}

func (v *UserListView) GetPermissions() []drf.Permission {
    return []drf.Permission{drf.IsAuthenticated()}
}

func (v *UserListView) Get(c *drf.Context) drf.Response {
    users, _ := orm.Objects[User]().All()
    data, _ := drf.SerializeMany[UserSerializer](users)
    return drf.OK(data)
}

func (v *UserListView) Post(c *drf.Context) drf.Response {
    serializer := &UserSerializer{}
    if err := serializer.Validate(c.Data); err != nil {
        return drf.BadRequest(serializer.Errors())
    }
    user, _ := serializer.Create(serializer.ValidatedData())
    data, _ := drf.Serialize[UserSerializer](user)
    return drf.Created(data)
}
```

### 7.2 ViewSet

**DRF:**
```python
from rest_framework import viewsets
from rest_framework.decorators import action
from rest_framework.response import Response

class UserViewSet(viewsets.ModelViewSet):
    queryset = User.objects.all()
    serializer_class = UserSerializer
    permission_classes = [IsAuthenticated]
    
    @action(detail=True, methods=['post'])
    def activate(self, request, pk=None):
        user = self.get_object()
        user.is_active = True
        user.save()
        return Response(UserSerializer(user).data)
```

**Django-DRF-Go:**
```go
type UserViewSet struct {
    drf.ModelViewSet[User]
}

func (v *UserViewSet) Meta() drf.ViewSetMeta {
    return drf.ViewSetMeta{
        QuerySet:    orm.Objects[User](),
        Serializer:  &UserSerializer{},
        Permissions: []drf.Permission{drf.IsAuthenticated()},
    }
}

func (v *UserViewSet) Actions() []drf.Action {
    return []drf.Action{
        {Name: "activate", Methods: []string{"POST"}, Detail: true, Handler: v.Activate},
    }
}

func (v *UserViewSet) Activate(c *drf.Context) drf.Response {
    user := v.GetObject(c)
    user.IsActive = true
    user.Save()
    return drf.OK(v.Serialize(user))
}
```

---

## 8. Per-Action Configuration

```go
func (v *UserViewSet) GetPermissions(action string) []drf.Permission {
    switch action {
    case "create", "update", "destroy":
        return []drf.Permission{drf.IsAdminUser()}
    case "activate", "deactivate":
        return []drf.Permission{drf.IsStaff()}
    default:
        return []drf.Permission{drf.IsAuthenticated()}
    }
}

func (v *UserViewSet) GetSerializer(action string) drf.Serializer {
    switch action {
    case "create":
        return &CreateUserSerializer{}
    case "update", "partial_update":
        return &UpdateUserSerializer{}
    case "list":
        return &UserListSerializer{}  // Minimal fields
    default:
        return &UserSerializer{}
    }
}
```

---

## 9. Request Parsing

```go
// Automatic content negotiation
func (v *UserViewSet) Create(c *drf.Context) drf.Response {
    // c.Data is automatically parsed from:
    // - application/json
    // - application/x-www-form-urlencoded
    // - multipart/form-data
    
    data := c.Data  // map[string]interface{}
}

// File uploads
func (v *DocumentViewSet) Create(c *drf.Context) drf.Response {
    file := c.Request.MultipartForm.File["document"][0]
    // Process file...
}
```

---

## 10. Response Rendering

```go
// Automatic content negotiation for response
return drf.OK(data)  // Renders as JSON by default

// Force specific format
return drf.OK(data).WithContentType("application/xml")

// Custom headers
return drf.Created(data).WithHeader("Location", url)
```

---

## 11. Related PRDs

- [14_PRD_DRF_Serializers.md](./14_PRD_DRF_Serializers.md) - Serializers
- [16_PRD_DRF_Routers.md](./16_PRD_DRF_Routers.md) - URL routing for ViewSets
- [17_PRD_DRF_Authentication.md](./17_PRD_DRF_Authentication.md) - Authentication
- [18_PRD_DRF_Permissions.md](./18_PRD_DRF_Permissions.md) - Permissions
