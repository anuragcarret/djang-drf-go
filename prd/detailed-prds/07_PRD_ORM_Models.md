# PRD: ORM Models

> **Module:** `orm/models`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **Django Equivalent:** `django.db.models.Model`

---

## 1. Overview

### 1.1 Purpose

The Models System provides the foundation for data persistence:

- **Struct-Based Models**: Go structs as database tables
- **Type-Safe Fields**: Compile-time field validation
- **Meta Configuration**: Table names, indexes, constraints
- **Instance Methods**: Save, delete, refresh operations
- **Lifecycle Hooks**: Pre/post save, delete callbacks

### 1.2 Design Philosophy

| Django Approach | Go-Idiomatic Approach |
|-----------------|----------------------|
| `Model` class with metaclass | Struct embedding `Model` type |
| `class Meta` nested class | `Meta()` method returning config |
| Dynamic field discovery | Struct tags + reflection |
| `objects` manager | Type-safe generic manager |
| `save()`/`delete()` methods | Interface methods on models |

---

## 2. Core Concepts

### 2.1 Base Model

```go
// Model is embedded in all ORM models
type Model struct {
    ID        uint64    `drf:"primary_key;auto_increment"`
    CreatedAt time.Time `drf:"auto_now_add"`
    UpdatedAt time.Time `drf:"auto_now"`
}

// ModelInterface defines required model methods
type ModelInterface interface {
    TableName() string
    Meta() *ModelMeta
    PrimaryKey() interface{}
    IsNew() bool
}
```

### 2.2 Model Definition

```go
// User model example
type User struct {
    orm.Model  // Embeds ID, CreatedAt, UpdatedAt
    
    Email     string `drf:"unique;max_length=255"`
    Username  string `drf:"unique;max_length=150;index"`
    Password  string `drf:"max_length=128"`
    FirstName string `drf:"max_length=30;blank"`
    LastName  string `drf:"max_length=30;blank"`
    IsActive  bool   `drf:"default=true"`
    IsStaff   bool   `drf:"default=false"`
    LastLogin *time.Time `drf:"null"`
}

// TableName returns the database table name
func (u *User) TableName() string {
    return "auth_user"  // Default would be "users"
}

// Meta returns model configuration
func (u *User) Meta() *orm.ModelMeta {
    return &orm.ModelMeta{
        Verbose:      "User",
        VerbosePlural: "Users",
        Ordering:     []string{"-created_at"},
        Indexes: []orm.Index{
            {Fields: []string{"email"}},
            {Fields: []string{"username", "is_active"}},
        },
    }
}
```

### 2.3 Model Meta

```go
// ModelMeta holds model configuration
type ModelMeta struct {
    // Verbose name for display
    Verbose string
    
    // VerbosePlural for plural display
    VerbosePlural string
    
    // Ordering default order fields
    Ordering []string
    
    // Indexes additional database indexes
    Indexes []Index
    
    // Constraints check constraints
    Constraints []Constraint
    
    // UniqueTogetherFields composite unique constraints
    UniqueTogether [][]string
    
    // Abstract models aren't created as tables
    Abstract bool
    
    // Managed false means no migration management
    Managed bool
    
    // DBTable overrides default table name
    DBTable string
    
    // App label for grouping
    AppLabel string
}

// Index defines a database index
type Index struct {
    Name   string
    Fields []string
    Unique bool
    Where  string  // Partial index condition
}

// Constraint defines a check constraint
type Constraint struct {
    Name  string
    Check string  // SQL expression
}
```

---

## 3. Field Tags

### 3.1 Tag Syntax

```go
`drf:"option1;option2=value;option3"`
```

### 3.2 Common Field Options

| Tag | Description | Example |
|-----|-------------|---------|
| `primary_key` | Set as primary key | `drf:"primary_key"` |
| `unique` | Unique constraint | `drf:"unique"` |
| `null` | Allow NULL | `drf:"null"` |
| `blank` | Allow empty string | `drf:"blank"` |
| `default` | Default value | `drf:"default=true"` |
| `max_length` | Maximum length | `drf:"max_length=255"` |
| `index` | Create index | `drf:"index"` |
| `db_column` | Custom column name | `drf:"db_column=user_email"` |
| `db_index` | Create database index | `drf:"db_index"` |
| `auto_now` | Update on save | `drf:"auto_now"` |
| `auto_now_add` | Set on create | `drf:"auto_now_add"` |
| `choices` | Limit to choices | `drf:"choices=draft,published"` |
| `validators` | Validation rules | `drf:"validators=email"` |

### 3.3 Relationship Tags

| Tag | Description | Example |
|-----|-------------|---------|
| `fk` | Foreign key | `drf:"fk=users.User"` |
| `on_delete` | Delete behavior | `drf:"on_delete=CASCADE"` |
| `related_name` | Reverse accessor | `drf:"related_name=posts"` |
| `m2m` | Many-to-many | `drf:"m2m=tags.Tag"` |
| `through` | Through table | `drf:"through=PostTag"` |
| `o2o` | One-to-one | `drf:"o2o=profiles.Profile"` |

---

## 4. API Specification

### 4.1 Model Instance Methods

```go
// Save persists the model to database
func (m *Model) Save(opts ...SaveOption) error

// SaveFields updates only specified fields
func (m *Model) SaveFields(fields ...string) error

// Delete removes the model from database
func (m *Model) Delete() error

// Refresh reloads from database
func (m *Model) Refresh() error

// GetOrCreate finds or creates record
func (m *Model) GetOrCreate(defaults map[string]interface{}) (bool, error)

// UpdateOrCreate updates or creates record
func (m *Model) UpdateOrCreate(defaults map[string]interface{}) (bool, error)
```

### 4.2 Save Options

```go
type SaveOption func(*saveConfig)

// ForceInsert always INSERT, never UPDATE
func ForceInsert() SaveOption

// ForceUpdate always UPDATE, never INSERT
func ForceUpdate() SaveOption

// UpdateFields limits fields to update
func UpdateFields(fields ...string) SaveOption

// Using specifies database alias
func Using(db string) SaveOption
```

### 4.3 Manager Access

```go
// Objects returns the default manager
func Objects[T ModelInterface]() *Manager[T]

// Usage:
users := orm.Objects[User]()
user, err := users.Get(orm.Q{"id": 1})
allUsers, err := users.All()
```

---

## 5. Django Comparison

### 5.1 Model Definition

**Django:**
```python
from django.db import models

class User(models.Model):
    email = models.EmailField(unique=True, max_length=255)
    username = models.CharField(unique=True, max_length=150, db_index=True)
    password = models.CharField(max_length=128)
    first_name = models.CharField(max_length=30, blank=True)
    last_name = models.CharField(max_length=30, blank=True)
    is_active = models.BooleanField(default=True)
    is_staff = models.BooleanField(default=False)
    last_login = models.DateTimeField(null=True, blank=True)
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)
    
    class Meta:
        ordering = ['-created_at']
        indexes = [
            models.Index(fields=['email']),
            models.Index(fields=['username', 'is_active']),
        ]
```

**Django-DRF-Go:**
```go
type User struct {
    orm.Model
    
    Email     string `drf:"unique;max_length=255;validators=email"`
    Username  string `drf:"unique;max_length=150;index"`
    Password  string `drf:"max_length=128"`
    FirstName string `drf:"max_length=30;blank"`
    LastName  string `drf:"max_length=30;blank"`
    IsActive  bool   `drf:"default=true"`
    IsStaff   bool   `drf:"default=false"`
    LastLogin *time.Time `drf:"null;blank"`
}

func (u *User) Meta() *orm.ModelMeta {
    return &orm.ModelMeta{
        Ordering: []string{"-created_at"},
        Indexes: []orm.Index{
            {Fields: []string{"email"}},
            {Fields: []string{"username", "is_active"}},
        },
    }
}
```

### 5.2 CRUD Operations

**Django:**
```python
# Create
user = User(email="john@example.com", username="john")
user.save()

# or
user = User.objects.create(email="john@example.com", username="john")

# Read
user = User.objects.get(id=1)

# Update
user.email = "new@example.com"
user.save()

# or partial update
user.save(update_fields=['email'])

# Delete
user.delete()
```

**Django-DRF-Go:**
```go
// Create
user := &User{Email: "john@example.com", Username: "john"}
err := user.Save()

// or
user, err := orm.Objects[User]().Create(orm.Q{
    "email": "john@example.com",
    "username": "john",
})

// Read
user, err := orm.Objects[User]().Get(orm.Q{"id": 1})

// Update
user.Email = "new@example.com"
err = user.Save()

// or partial update
err = user.SaveFields("email")

// Delete
err = user.Delete()
```

---

## 6. Model Inheritance

### 6.1 Abstract Base Models

```go
// TimestampedModel abstract base
type TimestampedModel struct {
    orm.Model
    CreatedAt time.Time `drf:"auto_now_add"`
    UpdatedAt time.Time `drf:"auto_now"`
}

func (t *TimestampedModel) Meta() *orm.ModelMeta {
    return &orm.ModelMeta{Abstract: true}
}

// Concrete model
type Article struct {
    TimestampedModel
    Title   string `drf:"max_length=200"`
    Content string `drf:"type=text"`
}
```

### 6.2 Multi-Table Inheritance

```go
// Base table
type Place struct {
    orm.Model
    Name    string `drf:"max_length=50"`
    Address string `drf:"max_length=80"`
}

// Child table with FK
type Restaurant struct {
    orm.Model
    Place        *Place `drf:"o2o=places.Place;parent"`
    ServesHotDogs bool
    ServesPizza   bool
}
```

### 6.3 Proxy Models

```go
// Proxy adds behavior without new table
type OrderedUser struct {
    User
}

func (o *OrderedUser) Meta() *orm.ModelMeta {
    m := o.User.Meta()
    m.Proxy = true
    m.Ordering = []string{"username"}
    return m
}
```

---

## 7. Custom Methods

### 7.1 Instance Methods

```go
type User struct {
    orm.Model
    FirstName string
    LastName  string
    Email     string
}

// FullName returns formatted name
func (u *User) FullName() string {
    return fmt.Sprintf("%s %s", u.FirstName, u.LastName)
}

// EmailDomain extracts domain from email
func (u *User) EmailDomain() string {
    parts := strings.Split(u.Email, "@")
    if len(parts) > 1 {
        return parts[1]
    }
    return ""
}

// IsEmailVerified checks verification status
func (u *User) IsEmailVerified() bool {
    return u.EmailVerifiedAt != nil
}
```

### 7.2 Computed Properties

```go
// String implements Stringer for display
func (u *User) String() string {
    return u.FullName()
}

// Age computes dynamic value (not stored)
func (u *User) Age() int {
    if u.BirthDate == nil {
        return 0
    }
    return time.Since(*u.BirthDate).Hours() / 24 / 365
}
```

---

## 8. Model Lifecycle Hooks

### 8.1 Hook Methods

```go
// PreSave called before save
func (u *User) PreSave() error {
    u.Email = strings.ToLower(u.Email)
    return nil
}

// PostSave called after save
func (u *User) PostSave(created bool) error {
    if created {
        return u.SendWelcomeEmail()
    }
    return nil
}

// PreDelete called before delete
func (u *User) PreDelete() error {
    if u.IsProtected {
        return errors.New("cannot delete protected user")
    }
    return nil
}

// PostDelete called after delete
func (u *User) PostDelete() error {
    return cache.Delete("user:" + u.ID)
}

// Validate called before save
func (u *User) Validate() error {
    if !isValidEmail(u.Email) {
        return errors.New("invalid email format")
    }
    return nil
}
```

### 8.2 Hook Interfaces

```go
// PreSaver interface for pre-save hook
type PreSaver interface {
    PreSave() error
}

// PostSaver interface for post-save hook
type PostSaver interface {
    PostSave(created bool) error
}

// PreDeleter interface for pre-delete hook
type PreDeleter interface {
    PreDelete() error
}

// PostDeleter interface for post-delete hook
type PostDeleter interface {
    PostDelete() error
}

// Validator interface for validation
type Validator interface {
    Validate() error
}
```

---

## 9. Field Types Mapping

### 9.1 Go to SQL Type Mapping

| Go Type | Default SQL | Options |
|---------|-------------|---------|
| `string` | `VARCHAR(255)` | `max_length`, `type=text` |
| `int`, `int32` | `INTEGER` | |
| `int64` | `BIGINT` | |
| `uint64` | `BIGINT UNSIGNED` | |
| `float32` | `REAL` | |
| `float64` | `DOUBLE` | |
| `bool` | `BOOLEAN` | |
| `time.Time` | `TIMESTAMP` | `type=date`, `type=time` |
| `[]byte` | `BYTEA/BLOB` | |
| `*T` | Nullable column | |
| Custom types | JSON | `type=json` |

### 9.2 Custom Field Types

```go
// JSON field for arbitrary data
type JSONField map[string]interface{}

// Implement sql.Scanner and driver.Valuer
func (j *JSONField) Scan(value interface{}) error
func (j JSONField) Value() (driver.Value, error)

// Usage
type Settings struct {
    orm.Model
    Preferences JSONField `drf:"type=jsonb"`
}
```

---

## 10. Validation

### 10.1 Built-in Validators

```go
// Field-level validation via tags
type User struct {
    Email    string `drf:"validators=email"`
    Age      int    `drf:"validators=min:0,max:150"`
    Username string `drf:"validators=regex:^[a-z]+$"`
    Website  string `drf:"validators=url;blank"`
}
```

### 10.2 Custom Validators

```go
// Register custom validator
orm.RegisterValidator("phone", func(value interface{}) error {
    s, ok := value.(string)
    if !ok {
        return errors.New("expected string")
    }
    if !phoneRegex.MatchString(s) {
        return errors.New("invalid phone format")
    }
    return nil
})

// Use in model
type Contact struct {
    Phone string `drf:"validators=phone"`
}
```

---

## 11. Model Registration

```go
// Register models with app
func (a *UsersApp) Ready() error {
    orm.Register[User]()
    orm.Register[Profile]()
    orm.Register[UserGroup]()
    return nil
}

// Or automatic registration via init
func init() {
    orm.Register[User]()
}
```

---

## 12. Implementation Notes

### 12.1 Reflection

- Model fields parsed via `reflect` at registration
- Field metadata cached for performance
- Type assertions for value conversions

### 12.2 Dependencies

| Package | Purpose |
|---------|---------|
| `reflect` | Struct introspection |
| `database/sql` | Database interface |
| `sync` | Thread-safe caches |

---

## 13. Related PRDs

- [08_PRD_ORM_Fields.md](./08_PRD_ORM_Fields.md) - Field types
- [09_PRD_ORM_QuerySet.md](./09_PRD_ORM_QuerySet.md) - Query building
- [10_PRD_ORM_Managers.md](./10_PRD_ORM_Managers.md) - Model managers
- [11_PRD_ORM_Migrations.md](./11_PRD_ORM_Migrations.md) - Schema migrations
- [12_PRD_ORM_Relationships.md](./12_PRD_ORM_Relationships.md) - Model relationships
