# PRD: DRF Serializers

> **Module:** `drf/serializers`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **DRF Equivalent:** `rest_framework.serializers`

---

## 1. Overview

### 1.1 Purpose

The Serializers System transforms data between Go structs and JSON:

- **Output Serialization**: Models/structs → JSON responses
- **Input Deserialization**: JSON requests → validated data
- **Validation**: Field-level and object-level validation
- **Nested Serializers**: Handle related objects

### 1.2 Design Philosophy

| DRF Approach | Go-Idiomatic Approach |
|--------------|----------------------|
| `Serializer` class | Struct with `json` tags |
| `ModelSerializer` | Auto-generated from models |
| `validate_<field>` | Field validators |
| `.is_valid()` / `.validated_data` | `Validate()` method |
| `to_representation` | `ToJSON()` method |

---

## 2. Core Concepts

### 2.1 Serializer Interface

```go
// Serializer transforms data for API representation
type Serializer interface {
    // Validate checks input data
    Validate(data map[string]interface{}) error
    
    // ValidatedData returns cleaned data after validation
    ValidatedData() map[string]interface{}
    
    // Data returns output representation
    Data() map[string]interface{}
    
    // Errors returns validation errors
    Errors() map[string][]string
}
```

### 2.2 Basic Serializer

```go
// Define serializer as struct with tags
type UserSerializer struct {
    ID        uint64    `json:"id" read_only:"true"`
    Email     string    `json:"email" required:"true" validators:"email"`
    Username  string    `json:"username" required:"true" min_length:"3" max_length:"150"`
    FirstName string    `json:"first_name" max_length:"30"`
    LastName  string    `json:"last_name" max_length:"30"`
    IsActive  bool      `json:"is_active" read_only:"true"`
    CreatedAt time.Time `json:"created_at" read_only:"true"`
}
```

### 2.3 Model Serializer

```go
// ModelSerializer auto-generates from model
type UserSerializer struct {
    drf.ModelSerializer[User]
    
    // Override or add fields
    FullName string `json:"full_name" read_only:"true" source:"FullName()"` 
}

// Meta configuration
func (s *UserSerializer) Meta() drf.SerializerMeta {
    return drf.SerializerMeta{
        Model:   &User{},
        Fields:  []string{"id", "email", "username", "first_name", "last_name", "is_active"},
        // Or exclude specific fields
        // Exclude: []string{"password"},
        ReadOnlyFields: []string{"id", "is_active", "created_at"},
    }
}
```

---

## 3. Field Types

### 3.1 Field Tags

| Tag | Description | Example |
|-----|-------------|---------|
| `json` | JSON field name | `json:"email"` |
| `required` | Field is required | `required:"true"` |
| `read_only` | Output only | `read_only:"true"` |
| `write_only` | Input only | `write_only:"true"` |
| `default` | Default value | `default:"pending"` |
| `source` | Model field/method | `source:"GetFullName()"` |
| `validators` | Validation rules | `validators:"email,min_length:5"` |
| `allow_null` | Allow null | `allow_null:"true"` |
| `allow_blank` | Allow empty string | `allow_blank:"true"` |
| `min_length` | Min string length | `min_length:"3"` |
| `max_length` | Max string length | `max_length:"100"` |
| `min_value` | Min numeric value | `min_value:"0"` |
| `max_value` | Max numeric value | `max_value:"100"` |
| `choices` | Allowed values | `choices:"draft,published"` |

### 3.2 Field Types

```go
// String fields
Email    string `json:"email" validators:"email"`
URL      string `json:"website" validators:"url"`
Slug     string `json:"slug" validators:"slug"`
UUID     string `json:"uuid" validators:"uuid"`
IP       string `json:"ip_address" validators:"ip"`

// Numeric fields
Age      int     `json:"age" min_value:"0" max_value:"150"`
Price    float64 `json:"price" min_value:"0"`
Quantity int     `json:"quantity" min_value:"1"`

// Boolean
IsActive bool `json:"is_active" default:"true"`

// DateTime
CreatedAt time.Time  `json:"created_at" format:"2006-01-02T15:04:05Z07:00"`
BirthDate *time.Time `json:"birth_date" format:"2006-01-02" allow_null:"true"`

// Choice
Status string `json:"status" choices:"draft,review,published"`

// File
Avatar string `json:"avatar" max_length:"200"` // URL to file
```

---

## 4. API Specification

### 4.1 Serialization (Output)

```go
// Serialize single object
func Serialize[S any](instance interface{}) (*S, error)

// Serialize multiple objects
func SerializeMany[S any](instances []interface{}) ([]*S, error)

// Usage:
user, _ := orm.Objects[User]().Get(orm.Q{"id": 1})
data, err := drf.Serialize[UserSerializer](user)
// Returns map[string]interface{} for JSON response

// Or use serializer directly
serializer := &UserSerializer{}
serializer.SetInstance(user)
jsonData := serializer.Data()
```

### 4.2 Deserialization (Input)

```go
// Parse and validate input
func Deserialize[S any](data map[string]interface{}) (*S, error)

// Usage:
serializer := &UserSerializer{}
err := serializer.Validate(requestData)
if err != nil {
    return serializer.Errors()  // Validation errors
}

// Get validated data
validated := serializer.ValidatedData()

// Create instance
user := &User{
    Email:    validated["email"].(string),
    Username: validated["username"].(string),
}
```

### 4.3 Create and Update

```go
// CreateSerializer for creating records
type CreateUserSerializer struct {
    drf.ModelSerializer[User]
    Password string `json:"password" write_only:"true" required:"true" min_length:"8"`
}

func (s *CreateUserSerializer) Create(validated map[string]interface{}) (*User, error) {
    user := &User{
        Email:    validated["email"].(string),
        Username: validated["username"].(string),
    }
    user.SetPassword(validated["password"].(string))
    err := user.Save()
    return user, err
}

// UpdateSerializer for updating records
func (s *UpdateUserSerializer) Update(instance *User, validated map[string]interface{}) (*User, error) {
    if email, ok := validated["email"]; ok {
        instance.Email = email.(string)
    }
    if username, ok := validated["username"]; ok {
        instance.Username = username.(string)
    }
    err := instance.Save()
    return instance, err
}
```

---

## 5. Validation

### 5.1 Field Validation

```go
type CreateUserSerializer struct {
    Email    string `json:"email" required:"true" validators:"email"`
    Username string `json:"username" required:"true" min_length:"3"`
    Age      int    `json:"age" min_value:"18" max_value:"120"`
}

// Built-in validators
// - email, url, slug, uuid, ip
// - min_length, max_length
// - min_value, max_value
// - regex:pattern
```

### 5.2 Custom Field Validators

```go
type CreateUserSerializer struct {
    drf.BaseSerializer
    
    Username string `json:"username"`
}

// ValidateUsername validates the username field
func (s *CreateUserSerializer) ValidateUsername(value string) (string, error) {
    // Check uniqueness
    exists, _ := orm.Objects[User]().Filter(orm.Q{"username": value}).Exists()
    if exists {
        return "", drf.ValidationError("username already taken")
    }
    
    // Transform
    return strings.ToLower(value), nil
}
```

### 5.3 Object-Level Validation

```go
// Validate performs cross-field validation
func (s *PasswordChangeSerializer) Validate(data map[string]interface{}) error {
    password := data["password"].(string)
    confirm := data["password_confirm"].(string)
    
    if password != confirm {
        return drf.ValidationError(map[string]string{
            "password_confirm": "Passwords do not match",
        })
    }
    
    return nil
}
```

---

## 6. Nested Serializers

### 6.1 Read-Only Nested

```go
type PostSerializer struct {
    drf.ModelSerializer[Post]
    
    Author UserSerializer `json:"author" read_only:"true"` // Nested serializer
}

// Output:
// {
//   "id": 1,
//   "title": "Hello World",
//   "author": {
//     "id": 1,
//     "username": "john"
//   }
// }
```

### 6.2 Writable Nested

```go
type OrderSerializer struct {
    drf.ModelSerializer[Order]
    
    Items []OrderItemSerializer `json:"items"` // Nested list
}

func (s *OrderSerializer) Create(validated map[string]interface{}) (*Order, error) {
    items := validated["items"].([]map[string]interface{})
    
    return orm.Transaction(func(tx *orm.DB) error {
        order := &Order{/* ... */}
        if err := tx.Create(order); err != nil {
            return nil, err
        }
        
        for _, itemData := range items {
            item := &OrderItem{
                OrderID:   order.ID,
                ProductID: itemData["product_id"].(uint64),
                Quantity:  itemData["quantity"].(int),
            }
            if err := tx.Create(item); err != nil {
                return nil, err
            }
        }
        
        return order, nil
    })
}
```

### 6.3 PrimaryKeyRelatedField

```go
type PostSerializer struct {
    AuthorID uint64 `json:"author" source:"author_id"` // FK as ID
}

// Input: {"author": 1}
// Validates that User with id=1 exists
```

### 6.4 SlugRelatedField

```go
type PostSerializer struct {
    CategorySlug string `json:"category" source:"category.slug" slug_field:"slug"`
}

// Input: {"category": "tech"}
// Looks up Category by slug
```

---

## 7. Django REST Framework Comparison

### 7.1 Basic Serializer

**DRF:**
```python
class UserSerializer(serializers.Serializer):
    id = serializers.IntegerField(read_only=True)
    email = serializers.EmailField()
    username = serializers.CharField(max_length=150)
    first_name = serializers.CharField(max_length=30, required=False)
    
    def validate_username(self, value):
        if User.objects.filter(username=value).exists():
            raise serializers.ValidationError("Username taken")
        return value.lower()
    
    def validate(self, data):
        # Cross-field validation
        return data
    
    def create(self, validated_data):
        return User.objects.create(**validated_data)
```

**Django-DRF-Go:**
```go
type UserSerializer struct {
    drf.BaseSerializer
    
    ID        uint64 `json:"id" read_only:"true"`
    Email     string `json:"email" required:"true" validators:"email"`
    Username  string `json:"username" required:"true" max_length:"150"`
    FirstName string `json:"first_name" max_length:"30"`
}

func (s *UserSerializer) ValidateUsername(value string) (string, error) {
    exists, _ := orm.Objects[User]().Filter(orm.Q{"username": value}).Exists()
    if exists {
        return "", drf.ValidationError("Username taken")
    }
    return strings.ToLower(value), nil
}

func (s *UserSerializer) Validate(data map[string]interface{}) error {
    // Cross-field validation
    return nil
}

func (s *UserSerializer) Create(validated map[string]interface{}) (*User, error) {
    user := &User{
        Email:    validated["email"].(string),
        Username: validated["username"].(string),
    }
    return user, user.Save()
}
```

### 7.2 ModelSerializer

**DRF:**
```python
class UserSerializer(serializers.ModelSerializer):
    full_name = serializers.SerializerMethodField()
    
    class Meta:
        model = User
        fields = ['id', 'email', 'username', 'full_name']
        read_only_fields = ['id']
    
    def get_full_name(self, obj):
        return f"{obj.first_name} {obj.last_name}"
```

**Django-DRF-Go:**
```go
type UserSerializer struct {
    drf.ModelSerializer[User]
    
    FullName string `json:"full_name" read_only:"true" source:"FullName()"`
}

func (s *UserSerializer) Meta() drf.SerializerMeta {
    return drf.SerializerMeta{
        Model:          &User{},
        Fields:         []string{"id", "email", "username", "full_name"},
        ReadOnlyFields: []string{"id"},
    }
}
```

---

## 8. Serializer Context

```go
// Context from request
type PostSerializer struct {
    drf.ModelSerializer[Post]
}

func (s *PostSerializer) Create(validated map[string]interface{}) (*Post, error) {
    // Access request context
    request := s.Context().Value("request").(*http.Request)
    user := drf.GetUser(request)
    
    post := &Post{
        Title:  validated["title"].(string),
        UserID: user.ID,  // Set from context
    }
    return post, post.Save()
}
```

---

## 9. Serializer Fields

### 9.1 Method Fields

```go
type UserSerializer struct {
    PostCount int `json:"post_count" read_only:"true" method:"GetPostCount"`
}

func (s *UserSerializer) GetPostCount(user *User) int {
    count, _ := orm.Objects[Post]().Filter(orm.Q{"user_id": user.ID}).Count()
    return int(count)
}
```

### 9.2 Computed Source

```go
type UserSerializer struct {
    // Source from model method
    FullName string `json:"full_name" source:"FullName()"`
    
    // Source from related field
    CompanyName string `json:"company_name" source:"company.name"`
}
```

---

## 10. Error Responses

```go
// Validation error format
{
    "email": ["This field is required."],
    "username": ["Username already taken.", "Must be at least 3 characters."],
    "password": ["Password too short."]
}

// Non-field errors
{
    "non_field_errors": ["Passwords do not match."]
}
```

---

## 11. Related PRDs

- [07_PRD_ORM_Models.md](./07_PRD_ORM_Models.md) - Model definitions
- [15_PRD_DRF_Views_ViewSets.md](./15_PRD_DRF_Views_ViewSets.md) - Using serializers
- [19_PRD_DRF_Filtering_Pagination.md](./19_PRD_DRF_Filtering_Pagination.md) - Pagination in responses
