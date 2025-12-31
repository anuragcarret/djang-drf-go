# PRD: DRF Filtering and Pagination

> **Module:** `drf/filtering`, `drf/pagination`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **DRF Equivalent:** `rest_framework.filters`, `rest_framework.pagination`

---

## 1. Overview

### 1.1 Purpose

Filter and paginate API responses:

- **Search**: Full-text search across fields
- **Filtering**: Filter by query parameters
- **Ordering**: Sort results
- **Pagination**: Page-based or cursor-based pagination

---

## 2. Filter Backends

### 2.1 SearchFilter

```go
drf.SearchFilter()

// Configure in ViewSet
func (v *UserViewSet) Meta() drf.ViewSetMeta {
    return drf.ViewSetMeta{
        SearchFields: []string{"username", "email", "first_name"},
    }
}

// Query: ?search=john
// Searches: username ILIKE '%john%' OR email ILIKE '%john%' OR ...
```

### 2.2 OrderingFilter

```go
drf.OrderingFilter()

func (v *UserViewSet) Meta() drf.ViewSetMeta {
    return drf.ViewSetMeta{
        OrderingFields: []string{"created_at", "username"},
        DefaultOrdering: []string{"-created_at"},
    }
}

// Query: ?ordering=-created_at,username
```

### 2.3 FieldFilter (DjangoFilterBackend)

```go
drf.FieldFilter()

func (v *UserViewSet) Meta() drf.ViewSetMeta {
    return drf.ViewSetMeta{
        FilterFields: map[string][]string{
            "is_active": {"exact"},
            "created_at": {"gte", "lte", "exact"},
            "username": {"exact", "icontains"},
        },
    }
}

// Query: ?is_active=true&created_at__gte=2024-01-01
```

---

## 3. Pagination

### 3.1 PageNumberPagination

```go
drf.PageNumberPagination()

// Config
type PageNumberPaginationConfig struct {
    PageSize          int    `default:"20"`
    PageSizeQueryParam string `default:"page_size"`
    MaxPageSize       int    `default:"100"`
}

// Query: ?page=2&page_size=10

// Response:
{
    "count": 100,
    "next": "/api/users/?page=3",
    "previous": "/api/users/?page=1",
    "results": [...]
}
```

### 3.2 LimitOffsetPagination

```go
drf.LimitOffsetPagination()

// Query: ?limit=10&offset=20

// Response:
{
    "count": 100,
    "next": "/api/users/?limit=10&offset=30",
    "previous": "/api/users/?limit=10&offset=10",
    "results": [...]
}
```

### 3.3 CursorPagination

```go
drf.CursorPagination()

// Query: ?cursor=cD0yMDI0LTAxLTAx

// Response:
{
    "next": "/api/users/?cursor=cD0yMDI0LTAxLTAy",
    "previous": "/api/users/?cursor=cj0xJnA9MjAyNC0wMS0wMQ",
    "results": [...]
}

// Best for real-time data with consistent ordering
```

---

## 4. Configuration

```go
// Global defaults in settings
{
    "rest_framework": {
        "default_pagination": "PageNumberPagination",
        "page_size": 20,
        "max_page_size": 100
    }
}

// Per-ViewSet
func (v *UserViewSet) GetPagination() drf.Pagination {
    return &drf.LimitOffsetPagination{MaxLimit: 50}
}
```

---

## 5. DRF Comparison

**DRF:**
```python
class UserViewSet(viewsets.ModelViewSet):
    filter_backends = [SearchFilter, OrderingFilter]
    search_fields = ['username', 'email']
    ordering_fields = ['created_at']
    pagination_class = PageNumberPagination
```

**Django-DRF-Go:**
```go
func (v *UserViewSet) Meta() drf.ViewSetMeta {
    return drf.ViewSetMeta{
        FilterBackends:  []drf.FilterBackend{drf.SearchFilter(), drf.OrderingFilter()},
        SearchFields:    []string{"username", "email"},
        OrderingFields:  []string{"created_at"},
        Pagination:      &drf.PageNumberPagination{},
    }
}
```

---

## 6. Related PRDs

- [09_PRD_ORM_QuerySet.md](./09_PRD_ORM_QuerySet.md) - Query filtering
- [15_PRD_DRF_Views_ViewSets.md](./15_PRD_DRF_Views_ViewSets.md) - ViewSets
