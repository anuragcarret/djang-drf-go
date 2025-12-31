# PRD: Contrib Auth

> **Module:** `contrib/auth`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **Django Equivalent:** `django.contrib.auth`

---

## 1. Overview

### 1.1 Purpose

User authentication and authorization:

- **User Model**: Accounts with authentication
- **Authentication**: Login, logout, password management
- **Authorization**: Permissions and groups
- **Password Hashing**: Secure password storage

---

## 2. User Model

```go
// User is the default user model
type User struct {
    orm.Model
    
    Username     string     `drf:"max_length=150;unique"`
    Email        string     `drf:"max_length=254;unique"`
    Password     string     `drf:"max_length=128"`  // Hashed
    FirstName    string     `drf:"max_length=30;blank"`
    LastName     string     `drf:"max_length=150;blank"`
    IsActive     bool       `drf:"default=true"`
    IsStaff      bool       `drf:"default=false"`
    IsSuperuser  bool       `drf:"default=false"`
    LastLogin    *time.Time `drf:"null"`
    DateJoined   time.Time  `drf:"auto_now_add"`
}

// Set password with hashing
func (u *User) SetPassword(raw string) {
    u.Password = auth.MakePassword(raw)
}

// Check password
func (u *User) CheckPassword(raw string) bool {
    return auth.CheckPassword(raw, u.Password)
}

// Full name
func (u *User) FullName() string {
    return strings.TrimSpace(u.FirstName + " " + u.LastName)
}
```

---

## 3. Authentication Functions

```go
// Authenticate verifies credentials
func Authenticate(username, password string) (*User, error)

// Login creates session for user
func Login(w http.ResponseWriter, r *http.Request, user *User) error

// Logout destroys session
func Logout(w http.ResponseWriter, r *http.Request) error

// GetUser returns current user from request
func GetUser(r *http.Request) *User
```

---

## 4. Password Hashing

```go
// MakePassword hashes with bcrypt
func MakePassword(raw string) string

// CheckPassword verifies password
func CheckPassword(raw, hashed string) bool

// Password validation
func ValidatePassword(password string, user *User) error

// Validators
var PasswordValidators = []PasswordValidator{
    MinimumLengthValidator{MinLength: 8},
    CommonPasswordValidator{},
    NumericPasswordValidator{},
    UserAttributeSimilarityValidator{},
}
```

---

## 5. Permissions

```go
// Permission model
type Permission struct {
    orm.Model
    Name        string `drf:"max_length=255"`
    Codename    string `drf:"max_length=100"`  // e.g., "add_user"
    ContentType string `drf:"max_length=100"`  // e.g., "auth.user"
}

// Group model
type Group struct {
    orm.Model
    Name        string        `drf:"max_length=150;unique"`
    Permissions []*Permission `drf:"m2m"`
}

// User permissions
func (u *User) HasPerm(perm string) bool
func (u *User) HasPerms(perms []string) bool
func (u *User) GetAllPermissions() []string
```

---

## 6. Django Comparison

**Django:**
```python
from django.contrib.auth.models import User
from django.contrib.auth import authenticate, login, logout

user = authenticate(username='john', password='secret')
if user is not None:
    login(request, user)

if request.user.has_perm('app.add_article'):
    # Can add articles
```

**Django-DRF-Go:**
```go
user, err := auth.Authenticate("john", "secret")
if err == nil {
    auth.Login(w, r, user)
}

currentUser := auth.GetUser(r)
if currentUser.HasPerm("articles.add_article") {
    // Can add articles
}
```

---

## 7. Management Commands

```bash
# Create superuser
./manage createsuperuser

# Change password
./manage changepassword <username>
```

---

## 8. Related PRDs

- [17_PRD_DRF_Authentication.md](./17_PRD_DRF_Authentication.md) - API auth
- [18_PRD_DRF_Permissions.md](./18_PRD_DRF_Permissions.md) - API permissions
- [25_PRD_Contrib_Sessions.md](./25_PRD_Contrib_Sessions.md) - Sessions
