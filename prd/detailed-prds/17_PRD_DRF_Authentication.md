# PRD: DRF Authentication

> **Module:** `drf/authentication`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **DRF Equivalent:** `rest_framework.authentication`

---

## 1. Overview

### 1.1 Purpose

The Authentication System verifies user identity:

- **Token Authentication**: API tokens for stateless auth
- **Session Authentication**: Cookie-based sessions
- **JWT Authentication**: JSON Web Tokens
- **Custom Backends**: Pluggable authentication

---

## 2. Core Concepts

### 2.1 Authentication Interface

```go
// Authentication validates requests
type Authentication interface {
    Authenticate(r *http.Request) (*auth.User, interface{}, error)
}

// Returns:
// - user: Authenticated user (nil if not authenticated)
// - auth: Auth info (token, session, etc.)
// - error: Only for auth failures (not unauthenticated)
```

---

## 3. Built-in Authentication

### 3.1 TokenAuthentication

```go
// TokenAuthentication uses database tokens
type TokenAuthentication struct{}

// Header: Authorization: Token <key>

// Token model
type Token struct {
    Key       string    `drf:"primary_key;max_length=40"`
    UserID    uint64    `drf:"fk=auth.User;unique"`
    CreatedAt time.Time `drf:"auto_now_add"`
}

// Create token
token := auth.CreateToken(user)
```

### 3.2 JWTAuthentication

```go
// JWTAuthentication uses JSON Web Tokens
type JWTAuthentication struct {
    SecretKey []byte
    Algorithm string  // HS256, RS256
}

// Header: Authorization: Bearer <token>

// Generate tokens
access, refresh := jwt.GenerateTokenPair(user, claims)

// Claims
type Claims struct {
    UserID    uint64
    TokenType string  // "access" or "refresh"
    Exp       int64
}
```

### 3.3 SessionAuthentication

```go
// SessionAuthentication uses Django sessions
type SessionAuthentication struct{}

// Reads user from session cookie
// Useful for browser-based API access
```

### 3.4 BasicAuthentication

```go
// BasicAuthentication uses HTTP Basic Auth
type BasicAuthentication struct{}

// Header: Authorization: Basic base64(username:password)
// Use only over HTTPS!
```

---

## 4. Configuration

```go
// In settings
type RESTConfig struct {
    DefaultAuthentication []string `json:"default_authentication"`
}

// settings.json
{
    "rest_framework": {
        "default_authentication": [
            "TokenAuthentication",
            "SessionAuthentication"
        ]
    }
}

// Per-view override
func (v *UserViewSet) GetAuthentication() []drf.Authentication {
    return []drf.Authentication{
        &drf.JWTAuthentication{},
    }
}
```

---

## 5. DRF Comparison

**DRF:**
```python
REST_FRAMEWORK = {
    'DEFAULT_AUTHENTICATION_CLASSES': [
        'rest_framework.authentication.TokenAuthentication',
    ]
}
```

**Django-DRF-Go:**
```json
{
    "rest_framework": {
        "default_authentication": ["TokenAuthentication"]
    }
}
```

---

## 6. Token Management

```go
// Generate token for user
token, _ := auth.Token.Create(user)

// Revoke token
auth.Token.Delete(user)

// Refresh JWT
newAccess, _ := jwt.RefreshToken(refreshToken)
```

---

## 7. Related PRDs

- [03_PRD_Core_Middleware.md](./03_PRD_Core_Middleware.md) - Auth middleware
- [18_PRD_DRF_Permissions.md](./18_PRD_DRF_Permissions.md) - Permissions
- [24_PRD_Contrib_Auth.md](./24_PRD_Contrib_Auth.md) - User model
