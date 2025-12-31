# PRD: Admin ModelAdmin

> **Module:** `admin/options`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **Django Equivalent:** `django.contrib.admin.ModelAdmin`

---

## 1. Overview

### 1.1 Purpose

Customize model display and editing in admin:

- **List View**: Columns, filters, search
- **Detail View**: Form fields, layout
- **Validation**: Custom validation rules
- **Permissions**: Per-model access control

---

## 2. List View Options

```go
type ArticleAdmin struct {
    admin.ModelAdmin
}

func (a *ArticleAdmin) Configure() {
    // Columns shown in list
    a.ListDisplay = []string{"title", "author", "status", "published_at", "view_count"}
    
    // Filter sidebar
    a.ListFilter = []string{"status", "author", "created_at"}
    
    // Search in these fields
    a.SearchFields = []string{"title", "content", "author__username"}
    
    // Pagination
    a.ListPerPage = 25
    
    // Default ordering
    a.Ordering = []string{"-published_at"}
    
    // Date hierarchy navigation
    a.DateHierarchy = "published_at"
    
    // Editable in list view
    a.ListEditable = []string{"status"}
    
    // Display links (click to edit)
    a.ListDisplayLinks = []string{"title"}
}
```

---

## 3. Detail/Form Options

```go
func (a *ArticleAdmin) Configure() {
    // Fields in form
    a.Fields = []string{"title", "slug", "content", "author", "status"}
    
    // Or exclude specific fields
    a.Exclude = []string{"view_count"}
    
    // Non-editable fields
    a.ReadonlyFields = []string{"created_at", "updated_at", "view_count"}
    
    // Field grouping
    a.Fieldsets = []admin.Fieldset{
        {
            Name: "Content",
            Fields: []string{"title", "slug", "content"},
        },
        {
            Name: "Publishing",
            Fields: []string{"author", "status", "published_at"},
            Classes: []string{"collapse"},
        },
    }
    
    // Prepopulate slug from title
    a.PrepopulatedFields = map[string][]string{
        "slug": {"title"},
    }
}
```

---

## 4. Custom Methods

```go
// Custom list column
func (a *ArticleAdmin) WordCount(obj *Article) string {
    count := len(strings.Fields(obj.Content))
    return fmt.Sprintf("%d words", count)
}

// Usage
a.ListDisplay = []string{"title", "WordCount", "status"}

// Custom filters
func (a *ArticleAdmin) GetFilters() []admin.Filter {
    return []admin.Filter{
        &PublishedFilter{},  // Custom filter
    }
}
```

---

## 5. Permissions

```go
// Override permission checks
func (a *ArticleAdmin) HasAddPermission(c *admin.Context) bool {
    return c.User.HasPerm("articles.add_article")
}

func (a *ArticleAdmin) HasChangePermission(c *admin.Context, obj *Article) bool {
    if obj != nil && obj.AuthorID != c.User.ID {
        return c.User.IsStaff  // Only author or staff can edit
    }
    return true
}

func (a *ArticleAdmin) HasDeletePermission(c *admin.Context, obj *Article) bool {
    return c.User.IsSuperuser
}
```

---

## 6. Django Comparison

**Django:**
```python
@admin.register(Article)
class ArticleAdmin(admin.ModelAdmin):
    list_display = ['title', 'author', 'status', 'word_count']
    list_filter = ['status', 'author']
    search_fields = ['title', 'content']
    readonly_fields = ['view_count']
    prepopulated_fields = {'slug': ('title',)}
    
    def word_count(self, obj):
        return len(obj.content.split())
```

**Django-DRF-Go:**
```go
type ArticleAdmin struct {
    admin.ModelAdmin
}

func (a *ArticleAdmin) Configure() {
    a.ListDisplay = []string{"title", "author", "status", "WordCount"}
    a.ListFilter = []string{"status", "author"}
    a.SearchFields = []string{"title", "content"}
    a.ReadonlyFields = []string{"view_count"}
    a.PrepopulatedFields = map[string][]string{"slug": {"title"}}
}

func (a *ArticleAdmin) WordCount(obj *Article) int {
    return len(strings.Fields(obj.Content))
}
```

---

## 7. Related PRDs

- [21_PRD_Admin_Dashboard.md](./21_PRD_Admin_Dashboard.md) - Admin site
- [23_PRD_Admin_Actions_Inlines.md](./23_PRD_Admin_Actions_Inlines.md) - Actions/Inlines
