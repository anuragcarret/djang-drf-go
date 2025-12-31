# PRD: DRF Permissions

> **Module:** `drf/permissions`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **DRF Equivalent:** `rest_framework.permissions`

---

## 1. Overview

### 1.1 Purpose

The Permissions System controls access to resources:

- **View-Level Permissions**: Check before any action
- **Object-Level Permissions**: Check per-object access
- **Composable Permissions**: AND/OR combinations
- **Custom Permissions**: Application-specific rules

---

## 2. Permission Interface

```go
// Permission checks access
type Permission interface {
    HasPermission(c *Context) bool
    HasObjectPermission(c *Context, obj interface{}) bool
}
```

---

## 3. Built-in Permissions

### 3.1 AllowAny

```go
drf.AllowAny()  // No restrictions
```

### 3.2 IsAuthenticated

```go
drf.IsAuthenticated()  // Must be logged in
```

### 3.3 IsAdminUser

```go
drf.IsAdminUser()  // user.IsStaff == true
```

### 3.4 IsAuthenticatedOrReadOnly

```go
drf.IsAuthenticatedOrReadOnly()  // Auth for writes, allow reads
```

### 3.5 DjangoModelPermissions

```go
drf.DjangoModelPermissions()  // Based on model permissions
// GET: view_<model>
// POST: add_<model>
// PUT/PATCH: change_<model>
// DELETE: delete_<model>
```

---

## 4. Custom Permissions

```go
type IsOwnerPermission struct{}

func (p *IsOwnerPermission) HasPermission(c *drf.Context) bool {
    return c.User != nil  // Must be authenticated
}

func (p *IsOwnerPermission) HasObjectPermission(c *drf.Context, obj interface{}) bool {
    post := obj.(*Post)
    return post.UserID == c.User.ID  // Must own object
}

// Usage
func (v *PostViewSet) GetPermissions() []drf.Permission {
    return []drf.Permission{
        drf.IsAuthenticated(),
        &IsOwnerPermission{},
    }
}
```

---

## 5. Composing Permissions

```go
// AND (all must pass)
drf.And(drf.IsAuthenticated(), drf.IsAdminUser())

// OR (any must pass)
drf.Or(drf.IsAdminUser(), &IsOwnerPermission{})
```

---

## 6. Per-Action Permissions

```go
func (v *PostViewSet) GetPermissions(action string) []drf.Permission {
    switch action {
    case "create", "update", "destroy":
        return []drf.Permission{drf.IsAuthenticated(), &IsOwnerPermission{}}
    default:
        return []drf.Permission{drf.AllowAny()}
    }
}
```

---

## 7. DRF Comparison

**DRF:**
```python
class IsOwner(permissions.BasePermission):
    def has_object_permission(self, request, view, obj):
        return obj.user == request.user

class PostViewSet(viewsets.ModelViewSet):
    permission_classes = [IsAuthenticated, IsOwner]
```

**Django-DRF-Go:**
```go
type IsOwnerPermission struct{}

func (p *IsOwnerPermission) HasObjectPermission(c *drf.Context, obj interface{}) bool {
    return obj.(*Post).UserID == c.User.ID
}

func (v *PostViewSet) GetPermissions() []drf.Permission {
    return []drf.Permission{drf.IsAuthenticated(), &IsOwnerPermission{}}
}
```

---

## 8. Related PRDs

- [17_PRD_DRF_Authentication.md](./17_PRD_DRF_Authentication.md) - Authentication
- [15_PRD_DRF_Views_ViewSets.md](./15_PRD_DRF_Views_ViewSets.md) - ViewSets
