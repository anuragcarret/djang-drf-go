# PRD: Unified Admin Site System

> **Module:** `django_drf_go/admin`
> **Version:** 1.0.0
> **Status:** Approved
> **Django Equivalent:** `django.contrib.admin`

---

## 1. Overview

### 1.1 Purpose
To provide a ready-to-use, automatic administrative interface for managing application models. It reads model metadata to provide a quick, model-centric interface where trusted users can manage content on your site.

### 1.2 Key Features
- **Auto-Discovery**: Automatically find registered models in each app.
- **Model Registry**: Centralized registry `admin.site` for all models.
- **Full CRUD Support**: Create, Read, Update, and Delete capabilities for all registered models.
- **CRUD Views**: Auto-generated List, Detail, Create, Update, Delete views.
- **API-Driven**: The admin site will be an API that a frontend (React/Vue/HTML) can consume.
- **Customizable**: Allow overriding list displays, filters, and search fields via `ModelAdmin`.

---

## 2. Core Components

### 2.1 AdminSite
The main entry point. It manages the registry of models and their associated `ModelAdmin` classes.

```go
type AdminSite struct {
    registry map[reflect.Type]*ModelAdmin
}

func (s *AdminSite) Register(model interface{}, adminConfig *ModelAdmin)
func (s *AdminSite) URLs() *urls.Router
```

### 2.2 ModelAdmin
Configuration struct for a model's admin interface.

```go
type ModelAdmin struct {
    ListDisplay  []string // Fields to show in list view
    ListFilter   []string // Fields to filter by
    SearchFields []string // Fields to search
    Ordering     []string // Default ordering
}
```

### 2.3 Admin API Endpoints
The admin site will expose standard REST endpoints for each registered model.

- `GET /admin/`: List all registered apps and models.
- `GET /admin/{app}/{model}/`: List objects (paginated, filtered).
- `POST /admin/{app}/{model}/`: Create object.
- `GET /admin/{app}/{model}/{id}/`: Retrieve object details.
- `PUT /admin/{app}/{model}/{id}/`: Update object.
- `DELETE /admin/{app}/{model}/{id}/`: Delete object.

---

## 3. Implementation Details

### 3.1 Metadata Extraction
Use reflection to extract field names, types, and struct tags (`drf:"..."`) to build the API responses dynamically.

### 3.2 Generic Views
Implement a `BaseAdminView` that uses `ModelAdmin` configuration to construct `QuerySet` filters and output serializers dynamically.

### 3.3 Integration
apps can register models in their `urls.go` or a new `admin.go` file (Django style).
```go
// demo/accounts/admin.go
func init() {
    admin.Site.Register(&User{}, &admin.ModelAdmin{
        ListDisplay: []string{"Username", "Email", "IsStaff"},
    })
}
```

---

## 4. Security
- **IsAdminUser Permission**: All admin routes must be protected by `IsAdminUser` permission class.
