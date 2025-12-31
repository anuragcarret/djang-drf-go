# PRD: ORM Fields

> **Module:** `orm/fields`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **Django Equivalent:** `django.db.models.fields`

---

## 1. Overview

### 1.1 Purpose

The Fields System defines database column types and behaviors:

- **Type Mapping**: Go types to SQL column types
- **Validation**: Field-level data validation
- **Serialization**: Database read/write transformations
- **Options**: Constraints, defaults, indexes

### 1.2 Design Philosophy

| Django Approach | Go-Idiomatic Approach |
|-----------------|----------------------|
| `Field` class hierarchy | Go types + struct tags |
| `CharField`, `IntegerField`, etc. | `string`, `int` with options |
| Field instances on model class | Struct fields with tags |
| Dynamic field introspection | Compiled field registry |

---

## 2. Core Concepts

### 2.1 Field Types Overview

```go
// Go type + drf tag = Field definition
type User struct {
    orm.Model
    
    // String fields
    Name      string  `drf:"max_length=100"`           // VARCHAR(100)
    Bio       string  `drf:"type=text"`                // TEXT
    Email     string  `drf:"max_length=255;validators=email"` // VARCHAR + validation
    
    // Numeric fields
    Age       int     `drf:""`                         // INTEGER
    Balance   float64 `drf:"decimal=10,2"`             // DECIMAL(10,2)
    
    // Boolean
    IsActive  bool    `drf:"default=true"`             // BOOLEAN
    
    // DateTime
    BirthDate time.Time `drf:"type=date"`              // DATE
    JoinedAt  time.Time `drf:"auto_now_add"`           // TIMESTAMP
}
```

### 2.2 Field Interface

```go
// Field represents a database column definition
type Field interface {
    // Name returns the Go field name
    Name() string
    
    // Column returns the database column name
    Column() string
    
    // SQLType returns the SQL type string
    SQLType(dialect string) string
    
    // GoType returns the Go reflect.Type
    GoType() reflect.Type
    
    // Options returns field configuration
    Options() *FieldOptions
    
    // ToDatabase converts Go value to DB value
    ToDatabase(value interface{}) (interface{}, error)
    
    // FromDatabase converts DB value to Go value
    FromDatabase(value interface{}) (interface{}, error)
    
    // Validate checks if value is valid
    Validate(value interface{}) error
}
```

---

## 3. String Fields

### 3.1 CharField (String with max_length)

```go
// VARCHAR column with max length
Name string `drf:"max_length=100"`

// Options:
// - max_length: Required, maximum character length
// - blank: Allow empty string (default: false)
// - choices: Limit to specific values

type Profile struct {
    Status string `drf:"max_length=20;choices=draft,published,archived;default=draft"`
}
```

### 3.2 TextField

```go
// TEXT column for long content
Description string `drf:"type=text"`

// Options:
// - blank: Allow empty (default: false)
// No max_length (unlimited)
```

### 3.3 EmailField

```go
// VARCHAR with email validation
Email string `drf:"max_length=254;validators=email"`

// Automatically validates email format
```

### 3.4 URLField

```go
// VARCHAR with URL validation
Website string `drf:"max_length=200;validators=url"`
```

### 3.5 SlugField

```go
// VARCHAR with slug validation
Slug string `drf:"max_length=50;validators=slug;index;unique"`

// Only lowercase letters, numbers, hyphens, underscores
```

### 3.6 UUIDField

```go
// UUID column
UUID uuid.UUID `drf:"primary_key"`

// or string representation
UUID string `drf:"type=uuid;default=uuid_generate_v4()"`
```

---

## 4. Numeric Fields

### 4.1 IntegerField

```go
// INTEGER columns
SmallInt  int16 `drf:""`         // SMALLINT
RegularInt int  `drf:""`         // INTEGER
BigInt    int64 `drf:""`         // BIGINT

// Unsigned variants
PositiveInt uint `drf:""`        // INTEGER (with CHECK >= 0)
```

### 4.2 FloatField

```go
// Floating point columns
Float32 float32 `drf:""`         // REAL
Float64 float64 `drf:""`         // DOUBLE PRECISION
```

### 4.3 DecimalField

```go
// Fixed precision decimals
Price   float64 `drf:"decimal=10,2"`    // DECIMAL(10,2)
Balance float64 `drf:"decimal=15,4"`    // DECIMAL(15,4)

// Or use decimal type
import "github.com/shopspring/decimal"
Price decimal.Decimal `drf:"decimal=10,2"`
```

### 4.4 AutoField

```go
// Auto-incrementing primary keys
orm.Model  // Includes ID uint64 with auto_increment

// Custom auto fields
ID   int   `drf:"primary_key;auto_increment"`     // SERIAL
ID64 int64 `drf:"primary_key;auto_increment"`     // BIGSERIAL
```

---

## 5. Boolean Field

```go
// Boolean column
IsActive bool `drf:"default=true"`
IsPublic bool `drf:"default=false"`

// Nullable boolean (three-state)
IsVerified *bool `drf:"null"`
```

---

## 6. DateTime Fields

### 6.1 DateTimeField

```go
// TIMESTAMP column
CreatedAt time.Time `drf:"auto_now_add"`       // Set on create
UpdatedAt time.Time `drf:"auto_now"`           // Update on save
PublishedAt time.Time `drf:"null;blank"`       // Nullable
```

### 6.2 DateField

```go
// DATE column (no time component)
BirthDate time.Time `drf:"type=date"`
ExpiryDate time.Time `drf:"type=date;null"`
```

### 6.3 TimeField

```go
// TIME column
StartTime time.Time `drf:"type=time"`
```

### 6.4 DurationField

```go
// INTERVAL column
Duration time.Duration `drf:""`              // Stored as microseconds
Timeout  time.Duration `drf:"default=5m"`
```

---

## 7. Binary Fields

### 7.1 BinaryField

```go
// BYTEA/BLOB column
Data []byte `drf:""`
Hash []byte `drf:"max_length=32"`

// For files, consider using FileField reference instead
```

---

## 8. Choice Fields

### 8.1 String Choices

```go
// Enum-like string values
type Article struct {
    Status string `drf:"max_length=20;choices=draft,review,published;default=draft"`
}

// Or define constants
const (
    StatusDraft     = "draft"
    StatusReview    = "review"
    StatusPublished = "published"
)
```

### 8.2 Integer Choices

```go
// Integer-based choices
type Priority int

const (
    PriorityLow    Priority = 1
    PriorityMedium Priority = 2
    PriorityHigh   Priority = 3
)

type Task struct {
    Priority Priority `drf:"choices=1,2,3;default=1"`
}
```

### 8.3 Choices with Labels

```go
// Define choice labels for display
var StatusChoices = orm.Choices{
    {Value: "draft", Label: "Draft"},
    {Value: "review", Label: "Under Review"},
    {Value: "published", Label: "Published"},
}

func (a *Article) Meta() *orm.ModelMeta {
    return &orm.ModelMeta{
        FieldChoices: map[string]orm.Choices{
            "Status": StatusChoices,
        },
    }
}
```

---

## 9. JSON Fields

### 9.1 Native JSON

```go
// JSONB column (Postgres)
type Settings struct {
    Preferences map[string]interface{} `drf:"type=jsonb"`
}

// Typed JSON
type NotificationSettings struct {
    Email bool `json:"email"`
    Push  bool `json:"push"`
    SMS   bool `json:"sms"`
}

type User struct {
    Notifications NotificationSettings `drf:"type=jsonb"`
}
```

### 9.2 Array Fields (Postgres)

```go
// Array columns
Tags   []string `drf:"type=array"`           // TEXT[]
Scores []int    `drf:"type=array"`           // INTEGER[]
```

---

## 10. Special Fields

### 10.1 IPAddressField

```go
// IP address storage
IPAddress string `drf:"type=inet"`           // INET (Postgres)
// or
IP net.IP `drf:""`                           // Custom scanner
```

### 10.2 GenericIPAddressField

```go
// IPv4 or IPv6
IP string `drf:"validators=ip"`
```

### 10.3 FilePathField

```go
// File system path
Path string `drf:"max_length=255;validators=filepath"`
```

---

## 11. Django Comparison

### 11.1 Field Definitions

**Django:**
```python
class Product(models.Model):
    name = models.CharField(max_length=100)
    description = models.TextField(blank=True)
    price = models.DecimalField(max_digits=10, decimal_places=2)
    quantity = models.PositiveIntegerField(default=0)
    is_available = models.BooleanField(default=True)
    created_at = models.DateTimeField(auto_now_add=True)
    category = models.CharField(
        max_length=20,
        choices=[
            ('electronics', 'Electronics'),
            ('clothing', 'Clothing'),
            ('food', 'Food'),
        ],
        default='electronics'
    )
```

**Django-DRF-Go:**
```go
type Product struct {
    orm.Model
    
    Name        string  `drf:"max_length=100"`
    Description string  `drf:"type=text;blank"`
    Price       float64 `drf:"decimal=10,2"`
    Quantity    uint    `drf:"default=0"`
    IsAvailable bool    `drf:"default=true"`
    Category    string  `drf:"max_length=20;choices=electronics,clothing,food;default=electronics"`
}

var CategoryChoices = orm.Choices{
    {Value: "electronics", Label: "Electronics"},
    {Value: "clothing", Label: "Clothing"},
    {Value: "food", Label: "Food"},
}
```

---

## 12. Field Validation

### 12.1 Built-in Validators

| Validator | Description | Example |
|-----------|-------------|---------|
| `email` | Valid email format | `validators=email` |
| `url` | Valid URL | `validators=url` |
| `slug` | Slug format | `validators=slug` |
| `ip` | IPv4/IPv6 address | `validators=ip` |
| `regex` | Custom regex | `validators=regex:^\d+$` |
| `min` | Minimum value | `validators=min:0` |
| `max` | Maximum value | `validators=max:100` |
| `min_length` | Minimum length | `validators=min_length:3` |
| `max_length` | Maximum length | (use max_length option) |

### 12.2 Custom Field Validators

```go
// Register field validator
orm.RegisterValidator("phone", func(value interface{}) error {
    str, ok := value.(string)
    if !ok {
        return errors.New("expected string")
    }
    if !phoneRegex.MatchString(str) {
        return errors.New("invalid phone number format")
    }
    return nil
})

// Use in model
type Contact struct {
    Phone string `drf:"max_length=20;validators=phone"`
}
```

---

## 13. Default Values

### 13.1 Static Defaults

```go
IsActive bool   `drf:"default=true"`
Status   string `drf:"default=pending"`
Priority int    `drf:"default=1"`
```

### 13.2 Dynamic Defaults

```go
// Database function defaults
UUID   string    `drf:"default=uuid_generate_v4()"`
Token  string    `drf:"default=gen_random_bytes(32)"`
Now    time.Time `drf:"default=now()"`

// Or use Go function via PreSave hook
func (u *User) PreSave() error {
    if u.Token == "" {
        u.Token = generateToken()
    }
    return nil
}
```

---

## 14. Nullable Fields

```go
// Use pointer types for nullable fields
type User struct {
    Email       string     `drf:"unique"`          // NOT NULL
    Phone       *string    `drf:"null"`            // NULL allowed
    LastLogin   *time.Time `drf:"null"`            // NULL allowed
    DeletedAt   *time.Time `drf:"null;index"`      // Soft delete pattern
}

// sql.NullString alternative
import "database/sql"
MiddleName sql.NullString `drf:"max_length=50"`
```

---

## 15. Field Options Reference

| Option | Type | Description |
|--------|------|-------------|
| `primary_key` | bool | Mark as primary key |
| `unique` | bool | Unique constraint |
| `null` | bool | Allow NULL values |
| `blank` | bool | Allow empty strings |
| `db_column` | string | Custom column name |
| `db_index` | bool | Create index |
| `index` | bool | Create index (alias) |
| `default` | any | Default value |
| `max_length` | int | Max string length |
| `decimal` | string | Precision,scale for decimals |
| `type` | string | Override SQL type |
| `choices` | string | Comma-separated valid values |
| `validators` | string | Validation rules |
| `auto_now` | bool | Update timestamp on save |
| `auto_now_add` | bool | Set timestamp on create |
| `editable` | bool | Can be modified (default: true) |
| `help_text` | string | Field description |

---

## 16. Implementation Notes

### 16.1 Type Registry

```go
// Fields are registered by Go type
var fieldRegistry = map[reflect.Kind]FieldFactory{
    reflect.String:  NewStringField,
    reflect.Int:     NewIntField,
    reflect.Int64:   NewBigIntField,
    reflect.Float64: NewFloatField,
    reflect.Bool:    NewBoolField,
    // ...
}
```

### 16.2 Dependencies

| Package | Purpose |
|---------|---------|
| `reflect` | Type introspection |
| `regexp` | Validation patterns |
| `database/sql` | SQL types |

---

## 17. Related PRDs

- [07_PRD_ORM_Models.md](./07_PRD_ORM_Models.md) - Model definitions
- [09_PRD_ORM_QuerySet.md](./09_PRD_ORM_QuerySet.md) - Field in queries
- [14_PRD_DRF_Serializers.md](./14_PRD_DRF_Serializers.md) - Field serialization
