# PRD: ORM QuerySet

> **Module:** `orm/queryset`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **Django Equivalent:** `django.db.models.QuerySet`

---

## 1. Overview

### 1.1 Purpose

The QuerySet System provides a fluent, chainable query builder:

- **Lazy Evaluation**: Queries execute only when results are accessed
- **Chainable Methods**: Filter, order, annotate, aggregate
- **Type Safety**: Generic querysets for compile-time type checking
- **SQL Generation**: Database-agnostic query building

### 1.2 Design Philosophy

| Django Approach | Go-Idiomatic Approach |
|-----------------|----------------------|
| Dynamic `QuerySet` class | Generic `QuerySet[T]` |
| `filter(**kwargs)` | `Filter(Q{...})` with typed maps |
| Method chaining returns new QS | Immutable chainable methods |
| Lazy evaluation | Same - lazy until iteration |
| `Q` objects for complex queries | `Q{}` and `Or()`, `And()` |

---

## 2. Core Concepts

### 2.1 QuerySet Structure

```go
// QuerySet is a lazy, chainable query builder
type QuerySet[T ModelInterface] struct {
    model     *ModelMeta
    db        *DB
    
    // Query parts
    filters   []Filter
    excludes  []Filter
    ordering  []string
    limit     int
    offset    int
    
    // Select related
    selectRelated []string
    prefetchRelated []Prefetch
    
    // Annotations
    annotations map[string]Expression
    
    // Aggregations
    groupBy []string
    having  []Filter
    
    // Flags
    distinct bool
    forUpdate bool
}
```

### 2.2 Q Object (Query Params)

```go
// Q represents query parameters
type Q map[string]interface{}

// Usage:
users := orm.Objects[User]().Filter(orm.Q{
    "is_active": true,
    "age__gte":  18,
})
```

### 2.3 Lookups

```go
// Field lookups via double-underscore suffix
"field__lookup": value

// Available lookups:
"exact"      // = (default)
"iexact"     // case-insensitive =
"contains"   // LIKE %val%
"icontains"  // case-insensitive LIKE
"startswith" // LIKE val%
"endswith"   // LIKE %val
"in"         // IN (...)
"gt"         // >
"gte"        // >=
"lt"         // <
"lte"        // <=
"range"      // BETWEEN
"isnull"     // IS NULL / IS NOT NULL
"regex"      // REGEXP
"date"       // DATE(field)
"year"       // EXTRACT(YEAR)
"month"      // EXTRACT(MONTH)
```

---

## 3. API Specification

### 3.1 Basic Operations

#### Filter

```go
// Filter returns matching records
func (q *QuerySet[T]) Filter(params ...Q) *QuerySet[T]

// Examples:
users := orm.Objects[User]().Filter(orm.Q{"is_active": true})
users := orm.Objects[User]().Filter(orm.Q{"age__gte": 18, "age__lte": 65})
```

#### Exclude

```go
// Exclude returns non-matching records
func (q *QuerySet[T]) Exclude(params ...Q) *QuerySet[T]

// Example: Active users excluding admins
users := orm.Objects[User]().
    Filter(orm.Q{"is_active": true}).
    Exclude(orm.Q{"is_staff": true})
```

#### Get

```go
// Get returns single record or error
func (q *QuerySet[T]) Get(params ...Q) (*T, error)

// Errors: ErrDoesNotExist, ErrMultipleObjectsReturned
user, err := orm.Objects[User]().Get(orm.Q{"id": 1})
user, err := orm.Objects[User]().Filter(orm.Q{"email": email}).Get()
```

#### All

```go
// All returns all records as slice
func (q *QuerySet[T]) All() ([]*T, error)

users, err := orm.Objects[User]().All()
```

#### First / Last

```go
// First returns first record or nil
func (q *QuerySet[T]) First() (*T, error)

// Last returns last record or nil
func (q *QuerySet[T]) Last() (*T, error)

user, _ := orm.Objects[User]().OrderBy("created_at").First()
```

### 3.2 Ordering

```go
// OrderBy sets ordering (prefix with - for DESC)
func (q *QuerySet[T]) OrderBy(fields ...string) *QuerySet[T]

// Ascending
users.OrderBy("username")

// Descending (prefix with -)
users.OrderBy("-created_at")

// Multiple fields
users.OrderBy("-created_at", "username")

// Random ordering
users.OrderBy("?")
```

### 3.3 Limiting

```go
// Limit returns at most n records
func (q *QuerySet[T]) Limit(n int) *QuerySet[T]

// Offset skips first n records
func (q *QuerySet[T]) Offset(n int) *QuerySet[T]

// Pagination pattern
page := 2
pageSize := 20
users := orm.Objects[User]().
    OrderBy("-created_at").
    Offset((page - 1) * pageSize).
    Limit(pageSize)
```

### 3.4 Complex Queries

#### OR Conditions

```go
// Or combines with OR
func Or(queries ...Q) Filter

users := orm.Objects[User]().Filter(
    orm.Or(
        orm.Q{"is_staff": true},
        orm.Q{"is_superuser": true},
    ),
)
// WHERE is_staff = true OR is_superuser = true
```

#### AND Conditions

```go
// And explicitly combines with AND
func And(queries ...Q) Filter

users := orm.Objects[User]().Filter(
    orm.And(
        orm.Q{"is_active": true},
        orm.Or(
            orm.Q{"role": "admin"},
            orm.Q{"role": "moderator"},
        ),
    ),
)
// WHERE is_active = true AND (role = 'admin' OR role = 'moderator')
```

#### NOT Conditions

```go
// Not negates a query
func Not(q Q) Filter

users := orm.Objects[User]().Filter(
    orm.Not(orm.Q{"status": "banned"}),
)
// WHERE NOT (status = 'banned')
```

### 3.5 Aggregations

```go
// Count returns number of records
func (q *QuerySet[T]) Count() (int64, error)

count, _ := orm.Objects[User]().Filter(orm.Q{"is_active": true}).Count()

// Aggregate performs aggregations
func (q *QuerySet[T]) Aggregate(aggregates ...Aggregate) (map[string]interface{}, error)

result, _ := orm.Objects[Order]().Aggregate(
    orm.Sum("total").As("total_sales"),
    orm.Avg("total").As("avg_order"),
    orm.Max("total").As("max_order"),
    orm.Min("total").As("min_order"),
    orm.Count("id").As("order_count"),
)
// result["total_sales"] = 50000.00
```

### 3.6 Annotations

```go
// Annotate adds computed fields
func (q *QuerySet[T]) Annotate(annotations ...Annotation) *QuerySet[T]

users := orm.Objects[User]().
    Annotate(
        orm.Count("posts").As("post_count"),
        orm.Sum("posts__views").As("total_views"),
    ).
    Filter(orm.Q{"post_count__gt": 5})
```

### 3.7 Values

```go
// Values returns maps instead of models
func (q *QuerySet[T]) Values(fields ...string) ([]map[string]interface{}, error)

// ValuesFlat returns single field values
func (q *QuerySet[T]) ValuesFlat(field string) ([]interface{}, error)

// Get only specific fields
data, _ := orm.Objects[User]().Values("id", "email", "username")
// []map[string]interface{}{{"id": 1, "email": "...", "username": "..."}, ...}

// Get flat list of emails
emails, _ := orm.Objects[User]().ValuesFlat("email")
// []interface{}{"user1@example.com", "user2@example.com", ...}
```

### 3.8 Exists

```go
// Exists returns true if any record matches
func (q *QuerySet[T]) Exists() (bool, error)

exists, _ := orm.Objects[User]().Filter(orm.Q{"email": email}).Exists()
if exists {
    return errors.New("email already taken")
}
```

---

## 4. Related Data

### 4.1 Select Related (Eager Loading)

```go
// SelectRelated loads FK relations in single query
func (q *QuerySet[T]) SelectRelated(fields ...string) *QuerySet[T]

// Load post with author in one query
posts := orm.Objects[Post]().
    SelectRelated("author").
    All()

// Access without additional query
post.Author.Username
```

### 4.2 Prefetch Related

```go
// PrefetchRelated loads relations in separate query
func (q *QuerySet[T]) PrefetchRelated(prefetches ...Prefetch) *QuerySet[T]

// Load users with their posts
users := orm.Objects[User]().
    PrefetchRelated(orm.Prefetch("posts")).
    All()

// With filtered prefetch
users := orm.Objects[User]().
    PrefetchRelated(
        orm.Prefetch("posts").Filter(orm.Q{"is_published": true}),
    ).
    All()
```

---

## 5. Modification Operations

### 5.1 Create

```go
// Create inserts new record
func (q *QuerySet[T]) Create(data Q) (*T, error)

user, err := orm.Objects[User]().Create(orm.Q{
    "email": "new@example.com",
    "username": "newuser",
})
```

### 5.2 Update

```go
// Update modifies matching records
func (q *QuerySet[T]) Update(data Q) (int64, error)

count, err := orm.Objects[User]().
    Filter(orm.Q{"last_login__lt": thirtyDaysAgo}).
    Update(orm.Q{"is_active": false})
```

### 5.3 Delete

```go
// Delete removes matching records
func (q *QuerySet[T]) Delete() (int64, error)

count, err := orm.Objects[Session]().
    Filter(orm.Q{"expires_at__lt": time.Now()}).
    Delete()
```

### 5.4 Bulk Operations

```go
// BulkCreate inserts multiple records
func (q *QuerySet[T]) BulkCreate(objects []*T, opts ...BulkOption) error

users := []*User{
    {Email: "user1@example.com"},
    {Email: "user2@example.com"},
}
err := orm.Objects[User]().BulkCreate(users, orm.BatchSize(100))

// BulkUpdate updates specified fields on multiple records
func (q *QuerySet[T]) BulkUpdate(objects []*T, fields []string) error
```

### 5.5 Get or Create

```go
// GetOrCreate retrieves or creates record
func (q *QuerySet[T]) GetOrCreate(lookup Q, defaults Q) (*T, bool, error)

user, created, err := orm.Objects[User]().GetOrCreate(
    orm.Q{"email": email},           // Lookup
    orm.Q{"username": "default"},    // Defaults if creating
)

// UpdateOrCreate updates or creates
func (q *QuerySet[T]) UpdateOrCreate(lookup Q, defaults Q) (*T, bool, error)
```

---

## 6. Django Comparison

### 6.1 Basic Queries

**Django:**
```python
# Filter
users = User.objects.filter(is_active=True, age__gte=18)

# Exclude
users = User.objects.exclude(is_staff=True)

# Get single
user = User.objects.get(id=1)

# Chaining
users = User.objects.filter(is_active=True).exclude(is_banned=True).order_by('-created_at')[:10]
```

**Django-DRF-Go:**
```go
// Filter
users := orm.Objects[User]().Filter(orm.Q{"is_active": true, "age__gte": 18}).All()

// Exclude
users := orm.Objects[User]().Exclude(orm.Q{"is_staff": true}).All()

// Get single
user, err := orm.Objects[User]().Get(orm.Q{"id": 1})

// Chaining
users := orm.Objects[User]().
    Filter(orm.Q{"is_active": true}).
    Exclude(orm.Q{"is_banned": true}).
    OrderBy("-created_at").
    Limit(10).
    All()
```

### 6.2 Complex Queries

**Django:**
```python
from django.db.models import Q

# OR query
users = User.objects.filter(Q(is_staff=True) | Q(is_superuser=True))

# Complex
users = User.objects.filter(
    Q(role='admin') | Q(role='moderator'),
    is_active=True,
)
```

**Django-DRF-Go:**
```go
// OR query
users := orm.Objects[User]().Filter(
    orm.Or(
        orm.Q{"is_staff": true},
        orm.Q{"is_superuser": true},
    ),
).All()

// Complex
users := orm.Objects[User]().Filter(
    orm.Q{"is_active": true},
    orm.Or(
        orm.Q{"role": "admin"},
        orm.Q{"role": "moderator"},
    ),
).All()
```

### 6.3 Aggregations

**Django:**
```python
from django.db.models import Count, Sum, Avg

# Simple count
count = User.objects.filter(is_active=True).count()

# Aggregations
result = Order.objects.aggregate(
    total=Sum('amount'),
    avg_order=Avg('amount'),
    order_count=Count('id'),
)

# Annotate
users = User.objects.annotate(post_count=Count('posts'))
```

**Django-DRF-Go:**
```go
// Simple count
count, _ := orm.Objects[User]().Filter(orm.Q{"is_active": true}).Count()

// Aggregations
result, _ := orm.Objects[Order]().Aggregate(
    orm.Sum("amount").As("total"),
    orm.Avg("amount").As("avg_order"),
    orm.Count("id").As("order_count"),
)

// Annotate
users := orm.Objects[User]().
    Annotate(orm.Count("posts").As("post_count")).
    All()
```

---

## 7. Raw SQL

```go
// Raw executes raw SQL with model mapping
func Raw[T ModelInterface](sql string, args ...interface{}) *RawQuerySet[T]

users, err := orm.Raw[User](
    "SELECT * FROM users WHERE email LIKE ? AND created_at > ?",
    "%@example.com",
    startDate,
).All()

// RawSQL for complex expressions
users := orm.Objects[User]().
    Filter(orm.RawSQL("LOWER(email) = ?", strings.ToLower(email))).
    All()
```

---

## 8. Distinct

```go
// Distinct removes duplicate rows
func (q *QuerySet[T]) Distinct(fields ...string) *QuerySet[T]

// All distinct rows
unique := orm.Objects[User]().Distinct().All()

// Distinct on specific fields (Postgres)
unique := orm.Objects[User]().Distinct("country").All()
```

---

## 9. Iteration

```go
// Iterator for memory-efficient processing
func (q *QuerySet[T]) Iterator(batchSize int) <-chan *T

for user := range orm.Objects[User]().Iterator(100) {
    processUser(user)
}

// Each with callback
func (q *QuerySet[T]) Each(fn func(*T) error) error

orm.Objects[User]().Each(func(u *User) error {
    return sendEmail(u)
})
```

---

## 10. Locking

```go
// SelectForUpdate locks rows
func (q *QuerySet[T]) SelectForUpdate(opts ...LockOption) *QuerySet[T]

// In transaction
tx.Objects[Account]().
    Filter(orm.Q{"id": accountID}).
    SelectForUpdate().
    Get()

// Options
SelectForUpdate(orm.NoWait)      // NOWAIT
SelectForUpdate(orm.SkipLocked)  // SKIP LOCKED
```

---

## 11. F Expressions

```go
// F references another field
func F(field string) Expression

// Increment a field
orm.Objects[Product]().
    Filter(orm.Q{"id": 1}).
    Update(orm.Q{"stock": orm.F("stock").Sub(1)})
// UPDATE products SET stock = stock - 1 WHERE id = 1

// Compare fields
products := orm.Objects[Product]().
    Filter(orm.Q{"stock__lt": orm.F("reorder_level")}).
    All()
// WHERE stock < reorder_level
```

---

## 12. Performance

### 12.1 Query Optimization

- Use `SelectRelated` to avoid N+1 queries on FK
- Use `PrefetchRelated` for reverse relations and M2M
- Use `Values()` when you don't need full model
- Use `Only()` / `Defer()` for partial loading

```go
// Load only specific fields
users := orm.Objects[User]().Only("id", "email").All()

// Defer loading of expensive fields
users := orm.Objects[User]().Defer("bio", "avatar").All()
```

### 12.2 Explain

```go
// Explain query plan
plan, _ := orm.Objects[User]().
    Filter(orm.Q{"email__contains": "@example.com"}).
    Explain()

fmt.Println(plan)
// Seq Scan on users (cost=0.00..1.04 rows=1 width=...)
```

---

## 13. Related PRDs

- [07_PRD_ORM_Models.md](./07_PRD_ORM_Models.md) - Model definitions
- [10_PRD_ORM_Managers.md](./10_PRD_ORM_Managers.md) - Custom managers
- [12_PRD_ORM_Relationships.md](./12_PRD_ORM_Relationships.md) - Related queries
- [13_PRD_ORM_Transactions.md](./13_PRD_ORM_Transactions.md) - Transaction context
