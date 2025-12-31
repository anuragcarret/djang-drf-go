# PRD: Admin Dashboard

> **Module:** `admin/`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **Django Equivalent:** `django.contrib.admin`

---

## 1. Overview

### 1.1 Purpose

Auto-generated admin interface for managing data:

- **Dashboard**: Overview with statistics
- **Model CRUD**: Create, read, update, delete records
- **Search & Filter**: Find records quickly
- **Actions**: Bulk operations on records
- **Customizable**: Themes, layout, permissions

---

## 2. Core Concepts

### 2.1 AdminSite

```go
// AdminSite manages admin interface
type AdminSite struct {
    Name      string
    URLPrefix string
    Models    map[string]*ModelAdmin
}

// Default site
var Site = NewAdminSite("admin", "/admin")

// Register models
func init() {
    admin.Site.Register(&User{}, &UserAdmin{})
    admin.Site.Register(&Post{}, nil)  // Use default ModelAdmin
}
```

### 2.2 ModelAdmin

```go
// ModelAdmin customizes model display
type ModelAdmin struct {
    // Display settings
    ListDisplay     []string        // Columns in list view
    ListFilter      []string        // Filter sidebar
    SearchFields    []string        // Search box fields
    ListPerPage     int             // Pagination
    Ordering        []string        // Default order
    
    // Form settings
    Fields          []string        // Form fields
    Exclude         []string        // Hidden fields
    ReadonlyFields  []string        // Non-editable
    
    // Relationships
    Inlines         []InlineAdmin   // Nested models
}
```

---

## 3. Registration

```go
type UserAdmin struct {
    admin.ModelAdmin
}

func (a *UserAdmin) Configure() {
    a.ListDisplay = []string{"username", "email", "is_active", "created_at"}
    a.ListFilter = []string{"is_active", "is_staff", "created_at"}
    a.SearchFields = []string{"username", "email"}
    a.Ordering = []string{"-created_at"}
}

func init() {
    admin.Site.Register(&User{}, &UserAdmin{})
}
```

---

## 4. URL Setup

```go
func SetupURLs() *urls.Router {
    router := urls.NewRouter()
    
    // Include admin URLs
    router.Include("/admin", admin.Site.URLs(), "admin")
    
    return router
}

// Generated URLs:
// /admin/                  - Dashboard
// /admin/users/            - User list
// /admin/users/add/        - Add user
// /admin/users/{id}/       - Edit user
// /admin/users/{id}/delete - Delete user
```

---

## 5. Authentication

```go
// Admin requires staff user
func (s *AdminSite) Middleware() []middleware.Middleware {
    return []middleware.Middleware{
        middleware.AuthenticationMiddleware,
        admin.StaffRequiredMiddleware,
    }
}

// Custom login page at /admin/login/
```

---

## 6. Django Comparison

**Django:**
```python
from django.contrib import admin

@admin.register(User)
class UserAdmin(admin.ModelAdmin):
    list_display = ['username', 'email', 'is_active']
    list_filter = ['is_active', 'is_staff']
    search_fields = ['username', 'email']

# urls.py
urlpatterns = [
    path('admin/', admin.site.urls),
]
```

**Django-DRF-Go:**
```go
type UserAdmin struct {
    admin.ModelAdmin
}

func (a *UserAdmin) Configure() {
    a.ListDisplay = []string{"username", "email", "is_active"}
    a.ListFilter = []string{"is_active", "is_staff"}
    a.SearchFields = []string{"username", "email"}
}

func init() {
    admin.Site.Register(&User{}, &UserAdmin{})
}

// urls.go
router.Include("/admin", admin.Site.URLs(), "admin")
```

---

## 7. Related PRDs

- [22_PRD_Admin_ModelAdmin.md](./22_PRD_Admin_ModelAdmin.md) - Customization
- [23_PRD_Admin_Actions_Inlines.md](./23_PRD_Admin_Actions_Inlines.md) - Actions
- [24_PRD_Contrib_Auth.md](./24_PRD_Contrib_Auth.md) - User model
