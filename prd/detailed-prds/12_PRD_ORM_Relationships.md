# PRD: ORM Relationships

> **Module:** `orm/relationships`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **Django Equivalent:** `django.db.models` (ForeignKey, ManyToManyField, etc.)

---

## 1. Overview

### 1.1 Purpose

The Relationships System defines connections between models:

- **ForeignKey**: Many-to-one relationships
- **OneToOne**: One-to-one relationships
- **ManyToMany**: Many-to-many with join tables
- **Lazy Loading**: Load related data on demand
- **Eager Loading**: Load related data upfront

### 1.2 Design Philosophy

| Django Approach | Go-Idiomatic Approach |
|-----------------|----------------------|
| `ForeignKey` field | Struct field with `drf:"fk"` tag |
| `related_name` | Explicit reverse accessor method |
| Lazy by default | Eager loading via `SelectRelated` |
| `_set` suffix for reverse | Explicit `Related()` methods |

---

## 2. Core Concepts

### 2.1 ForeignKey (Many-to-One)

```go
// Post belongs to User
type Post struct {
    orm.Model
    Title    string `drf:"max_length=200"`
    Content  string `drf:"type=text"`
    
    // Foreign key to User
    UserID   uint64 `drf:"index"`                      // FK column
    User     *User  `drf:"fk=users.User;on_delete=CASCADE"` // Related object
}

// The UserID field stores the ID
// The User field is the loaded related object (nil until loaded)
```

### 2.2 OneToOne

```go
// Profile has one-to-one with User
type Profile struct {
    orm.Model
    Bio      string `drf:"type=text;blank"`
    Avatar   string `drf:"max_length=200;blank"`
    
    // One-to-one relationship
    UserID   uint64 `drf:"unique"`                     // Unique FK
    User     *User  `drf:"o2o=users.User;on_delete=CASCADE"`
}
```

### 2.3 ManyToMany

```go
// Article has many Tags, Tag has many Articles
type Article struct {
    orm.Model
    Title   string `drf:"max_length=200"`
    
    // Many-to-many (creates join table)
    Tags    []*Tag `drf:"m2m=tags.Tag"`
}

type Tag struct {
    orm.Model
    Name string `drf:"max_length=50;unique"`
    
    // Reverse side (optional explicit declaration)
    Articles []*Article `drf:"m2m=articles.Article;related_name=tags"`
}
```

### 2.4 Through Table (Explicit Join)

```go
// Explicit through model for extra fields
type Article struct {
    orm.Model
    Title   string 
    Authors []*Author `drf:"m2m=authors.Author;through=ArticleAuthor"`
}

type ArticleAuthor struct {
    orm.Model
    ArticleID  uint64  `drf:"fk=articles.Article"`
    AuthorID   uint64  `drf:"fk=authors.Author"`
    Role       string  `drf:"max_length=50"` // Extra field: author, editor, etc.
    Order      int     // Display order
}
```

---

## 3. On Delete Behavior

```go
// on_delete options
type OnDelete string

const (
    CASCADE     OnDelete = "CASCADE"    // Delete related
    PROTECT     OnDelete = "PROTECT"    // Prevent deletion
    SET_NULL    OnDelete = "SET_NULL"   // Set FK to NULL
    SET_DEFAULT OnDelete = "SET_DEFAULT" // Set to default
    DO_NOTHING  OnDelete = "DO_NOTHING" // No action
)

// Usage
UserID uint64 `drf:"fk=users.User;on_delete=SET_NULL;null"`
```

---

## 4. API Specification

### 4.1 Loading Related Objects

#### Lazy Loading

```go
// Access triggers query (if not already loaded)
post, _ := orm.Objects[Post]().Get(orm.Q{"id": 1})

// This triggers a query if User is nil
author := post.User  // Lazy load

// Or explicitly load
err := post.LoadRelated("User")
```

#### Eager Loading (SelectRelated)

```go
// FK and O2O relations: use SelectRelated (JOIN)
posts, _ := orm.Objects[Post]().
    SelectRelated("user").          // Single relation
    SelectRelated("user", "category"). // Multiple
    All()

// Access without additional query
for _, post := range posts {
    fmt.Println(post.User.Username)  // Already loaded
}
```

#### Eager Loading (PrefetchRelated)

```go
// M2M and reverse FK: use PrefetchRelated (separate query)
articles, _ := orm.Objects[Article]().
    PrefetchRelated(orm.Prefetch("tags")).
    All()

// With filtering
authors, _ := orm.Objects[Author]().
    PrefetchRelated(
        orm.Prefetch("articles").
            Filter(orm.Q{"is_published": true}).
            OrderBy("-published_at").
            Limit(5),
    ).
    All()
```

### 4.2 Reverse Relations

```go
// User has many Posts (reverse of Post.User FK)
type User struct {
    orm.Model
    Username string
}

// Access reverse relation
func (u *User) Posts() *orm.QuerySet[Post] {
    return orm.Objects[Post]().Filter(orm.Q{"user_id": u.ID})
}

// Usage
user, _ := orm.Objects[User]().Get(orm.Q{"id": 1})
userPosts, _ := user.Posts().OrderBy("-created_at").All()
recentPosts, _ := user.Posts().Filter(orm.Q{"created_at__gte": lastWeek}).All()
```

### 4.3 Querying Across Relations

```go
// Filter by related field using double underscore
posts, _ := orm.Objects[Post]().
    Filter(orm.Q{"user__username": "john"}).
    All()

// Multiple levels deep
comments, _ := orm.Objects[Comment]().
    Filter(orm.Q{"post__user__is_active": true}).
    All()

// Filter by related count
users, _ := orm.Objects[User]().
    Annotate(orm.Count("posts").As("post_count")).
    Filter(orm.Q{"post_count__gte": 10}).
    All()
```

### 4.4 ManyToMany Operations

```go
// Get related objects
article, _ := orm.Objects[Article]().Get(orm.Q{"id": 1})
tags, _ := article.Tags().All()

// Add relation
article.Tags().Add(tag1, tag2)

// Remove relation  
article.Tags().Remove(tag1)

// Clear all
article.Tags().Clear()

// Set (replace all)
article.Tags().Set(tag1, tag2, tag3)

// Check membership
exists, _ := article.Tags().Contains(tag1)
```

---

## 5. Django Comparison

### 5.1 ForeignKey

**Django:**
```python
class Post(models.Model):
    title = models.CharField(max_length=200)
    user = models.ForeignKey(User, on_delete=models.CASCADE, related_name='posts')

# Usage
post = Post.objects.select_related('user').get(id=1)
print(post.user.username)

# Reverse
user = User.objects.prefetch_related('posts').get(id=1)
for post in user.posts.all():
    print(post.title)
```

**Django-DRF-Go:**
```go
type Post struct {
    orm.Model
    Title  string `drf:"max_length=200"`
    UserID uint64 
    User   *User  `drf:"fk=users.User;on_delete=CASCADE"`
}

func (u *User) Posts() *orm.QuerySet[Post] {
    return orm.Objects[Post]().Filter(orm.Q{"user_id": u.ID})
}

// Usage
post, _ := orm.Objects[Post]().SelectRelated("user").Get(orm.Q{"id": 1})
fmt.Println(post.User.Username)

// Reverse
user, _ := orm.Objects[User]().Get(orm.Q{"id": 1})
posts, _ := user.Posts().All()
for _, post := range posts {
    fmt.Println(post.Title)
}
```

### 5.2 ManyToMany

**Django:**
```python
class Article(models.Model):
    title = models.CharField(max_length=200)
    tags = models.ManyToManyField(Tag, related_name='articles')

# Usage
article = Article.objects.prefetch_related('tags').get(id=1)
for tag in article.tags.all():
    print(tag.name)

# Modify
article.tags.add(tag1, tag2)
article.tags.remove(tag1)
article.tags.set([tag2, tag3])
```

**Django-DRF-Go:**
```go
type Article struct {
    orm.Model
    Title string `drf:"max_length=200"`
    Tags  []*Tag `drf:"m2m=tags.Tag"`
}

func (a *Article) Tags() *orm.M2MManager[Tag] {
    return orm.M2M[Tag](a, "tags")
}

// Usage
article, _ := orm.Objects[Article]().PrefetchRelated("tags").Get(orm.Q{"id": 1})
tags, _ := article.Tags().All()
for _, tag := range tags {
    fmt.Println(tag.Name)
}

// Modify
article.Tags().Add(tag1, tag2)
article.Tags().Remove(tag1)
article.Tags().Set(tag2, tag3)
```

---

## 6. Self-Referential Relationships

```go
// Tree structure (category -> subcategories)
type Category struct {
    orm.Model
    Name     string     `drf:"max_length=100"`
    ParentID *uint64    `drf:"null;index"`
    Parent   *Category  `drf:"fk=categories.Category;on_delete=CASCADE;null"`
}

// Get children
func (c *Category) Children() *orm.QuerySet[Category] {
    return orm.Objects[Category]().Filter(orm.Q{"parent_id": c.ID})
}

// Get ancestors
func (c *Category) Ancestors() ([]*Category, error) {
    var ancestors []*Category
    current := c
    for current.ParentID != nil {
        parent, err := orm.Objects[Category]().Get(orm.Q{"id": *current.ParentID})
        if err != nil {
            return nil, err
        }
        ancestors = append(ancestors, parent)
        current = parent
    }
    return ancestors, nil
}
```

---

## 7. Generic Relations

```go
// ContentType-like system for polymorphic relations
type Comment struct {
    orm.Model
    Content       string `drf:"type=text"`
    
    // Generic FK
    ContentType   string `drf:"max_length=100"` // "articles.Article", "photos.Photo"
    ObjectID      uint64 `drf:"index"`
}

// Helper to get related object
func (c *Comment) GetContentObject() (interface{}, error) {
    modelMeta, err := orm.GetModel(c.ContentType)
    if err != nil {
        return nil, err
    }
    return orm.GetByPK(modelMeta, c.ObjectID)
}

// Helper to set related object
func (c *Comment) SetContentObject(obj orm.ModelInterface) {
    c.ContentType = orm.ContentType(obj)
    c.ObjectID = obj.PrimaryKey().(uint64)
}
```

---

## 8. Relationship Managers

### 8.1 M2M Manager

```go
// M2MManager provides M2M operations
type M2MManager[T ModelInterface] struct {
    source      ModelInterface
    targetModel *ModelMeta
    throughModel *ModelMeta
}

// All returns all related objects
func (m *M2MManager[T]) All() ([]*T, error)

// Add creates relations
func (m *M2MManager[T]) Add(objects ...*T) error

// Remove deletes relations
func (m *M2MManager[T]) Remove(objects ...*T) error

// Clear removes all relations
func (m *M2MManager[T]) Clear() error

// Set replaces all relations
func (m *M2MManager[T]) Set(objects ...*T) error

// Contains checks if object is related
func (m *M2MManager[T]) Contains(obj *T) (bool, error)

// Count returns number of related objects
func (m *M2MManager[T]) Count() (int64, error)

// Filter returns filtered queryset
func (m *M2MManager[T]) Filter(q Q) *QuerySet[T]
```

### 8.2 Reverse FK Manager

```go
// RelatedManager for reverse FK access
type RelatedManager[T ModelInterface] struct {
    source      ModelInterface
    foreignKey  string
}

func (m *RelatedManager[T]) All() ([]*T, error)
func (m *RelatedManager[T]) Filter(q Q) *QuerySet[T]
func (m *RelatedManager[T]) Count() (int64, error)
func (m *RelatedManager[T]) Create(data Q) (*T, error)
```

---

## 9. Cascade Operations

```go
// Configuring cascade behavior
type Order struct {
    orm.Model
    UserID uint64 `drf:"fk=users.User;on_delete=PROTECT"`  // Prevent user deletion
}

type OrderItem struct {
    orm.Model
    OrderID uint64 `drf:"fk=orders.Order;on_delete=CASCADE"` // Delete with order
}

type AuditLog struct {
    orm.Model
    UserID *uint64 `drf:"fk=users.User;on_delete=SET_NULL;null"` // Keep log, null user
}
```

---

## 10. Performance Considerations

### 10.1 N+1 Problem

```go
// BAD: N+1 queries
posts, _ := orm.Objects[Post]().All()
for _, post := range posts {
    fmt.Println(post.User.Username)  // Query per post!
}

// GOOD: 1 query with JOIN
posts, _ := orm.Objects[Post]().SelectRelated("user").All()
for _, post := range posts {
    fmt.Println(post.User.Username)  // Already loaded
}

// GOOD: 2 queries (main + prefetch)
authors, _ := orm.Objects[Author]().PrefetchRelated("articles").All()
```

### 10.2 Deep Nesting

```go
// Select multiple levels
posts, _ := orm.Objects[Post]().
    SelectRelated("user", "category", "category__parent").
    All()

post.User.Username         // Loaded
post.Category.Name         // Loaded
post.Category.Parent.Name  // Loaded
```

---

## 11. Related PRDs

- [07_PRD_ORM_Models.md](./07_PRD_ORM_Models.md) - Model definitions
- [09_PRD_ORM_QuerySet.md](./09_PRD_ORM_QuerySet.md) - Relationship queries
- [14_PRD_DRF_Serializers.md](./14_PRD_DRF_Serializers.md) - Nested serializers
