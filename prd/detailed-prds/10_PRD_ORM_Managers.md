# PRD: ORM Managers

> **Module:** `orm/managers`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **Django Equivalent:** `django.db.models.Manager`

---

## 1. Overview

### 1.1 Purpose

The Managers System provides custom query interfaces for models:

- **Default Manager**: Standard `Objects` interface for all models
- **Custom Managers**: Model-specific query methods
- **Multiple Managers**: Different querysets for different uses
- **Manager Methods**: Reusable query patterns

### 1.2 Design Philosophy

| Django Approach | Go-Idiomatic Approach |
|-----------------|----------------------|
| `Manager` class | Generic `Manager[T]` struct |
| `objects = Manager()` | `Objects[T]()` function |
| Override `get_queryset()` | Custom `QuerySet()` method |
| Manager as class attribute | Manager via type registration |

---

## 2. Core Concepts

### 2.1 Default Manager

```go
// Manager is the default query interface
type Manager[T ModelInterface] struct {
    db    *DB
    model *ModelMeta
}

// Objects returns the default manager for a model
func Objects[T ModelInterface]() *Manager[T]

// Usage:
users, err := orm.Objects[User]().All()
user, err := orm.Objects[User]().Get(orm.Q{"id": 1})
```

### 2.2 Manager Interface

```go
// ManagerInterface defines manager contract
type ManagerInterface[T ModelInterface] interface {
    // QuerySet returns the base queryset
    QuerySet() *QuerySet[T]
    
    // Model returns model metadata
    Model() *ModelMeta
    
    // Database returns database connection
    DB() *DB
}
```

### 2.3 Default Manager Methods

```go
// All QuerySet methods are available on Manager
type Manager[T ModelInterface] struct {
    // ...
}

// All returns all records
func (m *Manager[T]) All() ([]*T, error)

// Get returns single record
func (m *Manager[T]) Get(params ...Q) (*T, error)

// Filter returns filtered queryset
func (m *Manager[T]) Filter(params ...Q) *QuerySet[T]

// Exclude returns excluded queryset
func (m *Manager[T]) Exclude(params ...Q) *QuerySet[T]

// Create inserts new record
func (m *Manager[T]) Create(data Q) (*T, error)

// Count returns record count
func (m *Manager[T]) Count() (int64, error)

// First returns first record
func (m *Manager[T]) First() (*T, error)

// ... all other QuerySet methods
```

---

## 3. Custom Managers

### 3.1 Defining Custom Manager

```go
// UserManager with custom methods
type UserManager struct {
    orm.Manager[User]
}

// NewUserManager creates the manager
func NewUserManager() *UserManager {
    return &UserManager{
        Manager: orm.NewManager[User](),
    }
}

// Active returns only active users
func (m *UserManager) Active() *orm.QuerySet[User] {
    return m.Filter(orm.Q{"is_active": true})
}

// Staff returns staff users
func (m *UserManager) Staff() *orm.QuerySet[User] {
    return m.Filter(orm.Q{"is_staff": true})
}

// ByEmail finds user by email
func (m *UserManager) ByEmail(email string) (*User, error) {
    return m.Get(orm.Q{"email": email})
}

// RecentlyActive returns users who logged in recently
func (m *UserManager) RecentlyActive(days int) *orm.QuerySet[User] {
    since := time.Now().AddDate(0, 0, -days)
    return m.Filter(orm.Q{"last_login__gte": since})
}
```

### 3.2 Registering Custom Manager

```go
// Option 1: Register as default
func init() {
    orm.RegisterManager[User](NewUserManager())
}

// Access via Objects
users := orm.Objects[User]().(*UserManager).Active().All()

// Option 2: Named manager
var Users = NewUserManager()

// Access directly
users := Users.Active().All()
activeStaff := Users.Staff().Filter(orm.Q{"is_active": true}).All()
```

### 3.3 Multiple Managers

```go
type Article struct {
    orm.Model
    Title       string
    Status      string
    PublishedAt *time.Time
}

// PublishedManager only returns published articles
type PublishedManager struct {
    orm.Manager[Article]
}

func (m *PublishedManager) QuerySet() *orm.QuerySet[Article] {
    return m.Manager.QuerySet().Filter(orm.Q{"status": "published"})
}

// Managers
var (
    Articles         = orm.NewManager[Article]()        // All articles
    PublishedArticles = &PublishedManager{Articles}     // Only published
)

// Usage
all, _ := Articles.All()                              // All articles
published, _ := PublishedArticles.All()               // Only published
drafts, _ := Articles.Filter(orm.Q{"status": "draft"}).All()
```

---

## 4. Manager Methods Pattern

### 4.1 Query Methods

```go
type OrderManager struct {
    orm.Manager[Order]
}

// Pending returns pending orders
func (m *OrderManager) Pending() *orm.QuerySet[Order] {
    return m.Filter(orm.Q{"status": "pending"})
}

// Completed returns completed orders  
func (m *OrderManager) Completed() *orm.QuerySet[Order] {
    return m.Filter(orm.Q{"status": "completed"})
}

// ForUser returns orders for a specific user
func (m *OrderManager) ForUser(userID uint64) *orm.QuerySet[Order] {
    return m.Filter(orm.Q{"user_id": userID})
}

// RecentOrders returns orders from last N days
func (m *OrderManager) Recent(days int) *orm.QuerySet[Order] {
    since := time.Now().AddDate(0, 0, -days)
    return m.Filter(orm.Q{"created_at__gte": since})
}

// WithItems eagerly loads order items
func (m *OrderManager) WithItems() *orm.QuerySet[Order] {
    return m.QuerySet().PrefetchRelated(orm.Prefetch("items"))
}
```

### 4.2 Aggregation Methods

```go
// TotalRevenue calculates total revenue
func (m *OrderManager) TotalRevenue() (float64, error) {
    result, err := m.Completed().Aggregate(orm.Sum("total"))
    if err != nil {
        return 0, err
    }
    return result["total__sum"].(float64), nil
}

// AverageOrderValue returns average order amount
func (m *OrderManager) AverageOrderValue() (float64, error) {
    result, err := m.Aggregate(orm.Avg("total"))
    if err != nil {
        return 0, err
    }
    return result["total__avg"].(float64), nil
}

// OrdersByStatus returns count per status
func (m *OrderManager) OrdersByStatus() (map[string]int64, error) {
    results, err := m.QuerySet().
        GroupBy("status").
        Annotate(orm.Count("id").As("count")).
        Values("status", "count")
    
    stats := make(map[string]int64)
    for _, r := range results {
        stats[r["status"].(string)] = r["count"].(int64)
    }
    return stats, err
}
```

### 4.3 Business Logic Methods

```go
// CreateWithItems creates order with items atomically
func (m *OrderManager) CreateWithItems(userID uint64, items []OrderItemData) (*Order, error) {
    var order *Order
    
    err := orm.Transaction(func(tx *orm.DB) error {
        var err error
        order, err = tx.Objects[Order]().Create(orm.Q{
            "user_id": userID,
            "status":  "pending",
            "total":   calculateTotal(items),
        })
        if err != nil {
            return err
        }
        
        for _, item := range items {
            _, err = tx.Objects[OrderItem]().Create(orm.Q{
                "order_id":   order.ID,
                "product_id": item.ProductID,
                "quantity":   item.Quantity,
                "price":      item.Price,
            })
            if err != nil {
                return err
            }
        }
        
        return nil
    })
    
    return order, err
}
```

---

## 5. Django Comparison

### 5.1 Custom Manager

**Django:**
```python
class UserManager(models.Manager):
    def active(self):
        return self.filter(is_active=True)
    
    def staff(self):
        return self.filter(is_staff=True)
    
    def by_email(self, email):
        return self.get(email=email)

class User(models.Model):
    # ...
    objects = UserManager()

# Usage
active_users = User.objects.active()
user = User.objects.by_email("john@example.com")
```

**Django-DRF-Go:**
```go
type UserManager struct {
    orm.Manager[User]
}

func (m *UserManager) Active() *orm.QuerySet[User] {
    return m.Filter(orm.Q{"is_active": true})
}

func (m *UserManager) Staff() *orm.QuerySet[User] {
    return m.Filter(orm.Q{"is_staff": true})
}

func (m *UserManager) ByEmail(email string) (*User, error) {
    return m.Get(orm.Q{"email": email})
}

var Users = &UserManager{orm.NewManager[User]()}

// Usage
activeUsers, _ := Users.Active().All()
user, _ := Users.ByEmail("john@example.com")
```

### 5.2 Filtered Manager

**Django:**
```python
class PublishedManager(models.Manager):
    def get_queryset(self):
        return super().get_queryset().filter(status='published')

class Article(models.Model):
    objects = models.Manager()  # Default
    published = PublishedManager()  # Filtered

# Usage
all_articles = Article.objects.all()
published = Article.published.all()
```

**Django-DRF-Go:**
```go
type PublishedManager struct {
    orm.Manager[Article]
}

func (m *PublishedManager) QuerySet() *orm.QuerySet[Article] {
    return m.Manager.QuerySet().Filter(orm.Q{"status": "published"})
}

var (
    Articles          = orm.NewManager[Article]()
    PublishedArticles = &PublishedManager{Articles}
)

// Usage
allArticles, _ := Articles.All()
published, _ := PublishedArticles.All()
```

---

## 6. Manager with Database Selection

```go
// Using specific database
func (m *UserManager) FromReplica() *orm.QuerySet[User] {
    return m.QuerySet().Using("replica")
}

func (m *UserManager) FromPrimary() *orm.QuerySet[User] {
    return m.QuerySet().Using("default")
}

// Usage
users, _ := Users.FromReplica().Filter(orm.Q{"is_active": true}).All()
```

---

## 7. Chainable Custom Methods

```go
// Return QuerySet for chaining
func (m *OrderManager) ForUser(userID uint64) *orm.QuerySet[Order] {
    return m.Filter(orm.Q{"user_id": userID})
}

func (m *OrderManager) Pending() *orm.QuerySet[Order] {
    return m.Filter(orm.Q{"status": "pending"})
}

// Chain custom methods with standard methods
orders, _ := Orders.ForUser(userID).Pending().OrderBy("-created_at").Limit(10).All()
```

---

## 8. Related Managers

```go
// Access related objects via manager
type User struct {
    orm.Model
    // ...
}

// RelatedManager for user's posts
func (u *User) Posts() *orm.QuerySet[Post] {
    return orm.Objects[Post]().Filter(orm.Q{"user_id": u.ID})
}

// Usage
user, _ := Users.Get(orm.Q{"id": 1})
userPosts, _ := user.Posts().OrderBy("-created_at").All()
recentPosts, _ := user.Posts().Filter(orm.Q{"created_at__gte": lastWeek}).All()
```

---

## 9. Testing Managers

```go
func TestUserManager(t *testing.T) {
    // Setup test database
    orm.SetupTestDB()
    defer orm.TeardownTestDB()
    
    // Create test data
    Users.Create(orm.Q{"email": "active@test.com", "is_active": true})
    Users.Create(orm.Q{"email": "inactive@test.com", "is_active": false})
    
    // Test custom manager method
    active, err := Users.Active().All()
    assert.NoError(t, err)
    assert.Len(t, active, 1)
    assert.Equal(t, "active@test.com", active[0].Email)
}
```

---

## 10. Implementation Notes

### 10.1 Manager Registration

```go
// Managers registered per model type
var managerRegistry = map[reflect.Type]interface{}{}

func RegisterManager[T ModelInterface](m ManagerInterface[T]) {
    managerRegistry[reflect.TypeOf((*T)(nil)).Elem()] = m
}

func Objects[T ModelInterface]() ManagerInterface[T] {
    t := reflect.TypeOf((*T)(nil)).Elem()
    if m, ok := managerRegistry[t]; ok {
        return m.(ManagerInterface[T])
    }
    return NewManager[T]()
}
```

### 10.2 Dependencies

| Package | Purpose |
|---------|---------|
| `reflect` | Type registration |
| `sync` | Thread-safe registry |

---

## 11. Related PRDs

- [07_PRD_ORM_Models.md](./07_PRD_ORM_Models.md) - Model definitions
- [09_PRD_ORM_QuerySet.md](./09_PRD_ORM_QuerySet.md) - QuerySet methods
- [12_PRD_ORM_Relationships.md](./12_PRD_ORM_Relationships.md) - Related managers
