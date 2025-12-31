# PRD: Admin Actions and Inlines

> **Module:** `admin/actions`, `admin/inlines`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **Django Equivalent:** `admin.action`, `admin.TabularInline`

---

## 1. Overview

### 1.1 Purpose

- **Actions**: Bulk operations on selected records
- **Inlines**: Edit related models on same page

---

## 2. Admin Actions

### 2.1 Defining Actions

```go
type ArticleAdmin struct {
    admin.ModelAdmin
}

func (a *ArticleAdmin) Configure() {
    a.Actions = []string{"publish", "unpublish", "make_featured"}
}

// Action: Publish selected articles
func (a *ArticleAdmin) Publish(c *admin.Context, queryset *orm.QuerySet[Article]) error {
    count, err := queryset.Update(orm.Q{"status": "published", "published_at": time.Now()})
    if err != nil {
        return err
    }
    c.Message(fmt.Sprintf("%d articles published", count))
    return nil
}

func (a *ArticleAdmin) Unpublish(c *admin.Context, queryset *orm.QuerySet[Article]) error {
    count, _ := queryset.Update(orm.Q{"status": "draft"})
    c.Message(fmt.Sprintf("%d articles unpublished", count))
    return nil
}
```

### 2.2 Action Descriptions

```go
func (a *ArticleAdmin) ActionDescriptions() map[string]string {
    return map[string]string{
        "publish":   "Mark selected articles as published",
        "unpublish": "Mark selected articles as draft",
    }
}
```

### 2.3 Confirmation

```go
func (a *ArticleAdmin) Delete(c *admin.Context, queryset *orm.QuerySet[Article]) error {
    // Require confirmation
    if !c.Confirmed {
        return admin.RequireConfirmation(
            "Delete articles",
            fmt.Sprintf("Delete %d selected articles?", queryset.Count()),
        )
    }
    
    queryset.Delete()
    return nil
}
```

---

## 3. Inline Admin

### 3.1 Tabular Inline

```go
type OrderItemInline struct {
    admin.TabularInline
    Model *OrderItem
}

func (i *OrderItemInline) Configure() {
    i.Fields = []string{"product", "quantity", "price"}
    i.Extra = 1  // Empty forms to show
    i.CanDelete = true
}

type OrderAdmin struct {
    admin.ModelAdmin
}

func (a *OrderAdmin) Configure() {
    a.Inlines = []admin.Inline{&OrderItemInline{}}
}
```

### 3.2 Stacked Inline

```go
type CommentInline struct {
    admin.StackedInline
    Model *Comment
}

func (i *CommentInline) Configure() {
    i.Fields = []string{"author", "content", "created_at"}
    i.ReadonlyFields = []string{"created_at"}
    i.Extra = 0
    i.MaxNum = 10
}
```

---

## 4. Django Comparison

**Django Actions:**
```python
@admin.action(description='Publish selected')
def publish(modeladmin, request, queryset):
    queryset.update(status='published')

class ArticleAdmin(admin.ModelAdmin):
    actions = [publish]
```

**Django-DRF-Go:**
```go
func (a *ArticleAdmin) Configure() {
    a.Actions = []string{"publish"}
}

func (a *ArticleAdmin) Publish(c *admin.Context, qs *orm.QuerySet[Article]) error {
    qs.Update(orm.Q{"status": "published"})
    return nil
}
```

**Django Inlines:**
```python
class OrderItemInline(admin.TabularInline):
    model = OrderItem
    extra = 1

class OrderAdmin(admin.ModelAdmin):
    inlines = [OrderItemInline]
```

**Django-DRF-Go:**
```go
type OrderItemInline struct {
    admin.TabularInline
    Model *OrderItem
}

func (a *OrderAdmin) Configure() {
    a.Inlines = []admin.Inline{&OrderItemInline{}}
}
```

---

## 5. Related PRDs

- [21_PRD_Admin_Dashboard.md](./21_PRD_Admin_Dashboard.md) - Admin site
- [22_PRD_Admin_ModelAdmin.md](./22_PRD_Admin_ModelAdmin.md) - ModelAdmin
