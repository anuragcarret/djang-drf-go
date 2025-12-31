# PRD: Contrib Sessions

> **Module:** `contrib/sessions`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **Django Equivalent:** `django.contrib.sessions`

---

## 1. Overview

### 1.1 Purpose

Server-side session management:

- **Session Storage**: Store user state between requests
- **Multiple Backends**: Database, cache, file, cookie
- **Security**: Secure session handling

---

## 2. Session Interface

```go
// Session stores user data
type Session interface {
    Get(key string) interface{}
    Set(key string, value interface{})
    Delete(key string)
    Clear()
    Save() error
    SessionKey() string
}

// Access from request
session := sessions.Get(r)
session.Set("cart_id", cartID)
session.Save()

// Read
cartID := session.Get("cart_id")
```

---

## 3. Session Backends

```go
// Database backend (default)
type DatabaseSessionStore struct{}

// Cache backend (Redis/Memcached)
type CacheSessionStore struct {
    Cache cache.Cache
}

// File backend
type FileSessionStore struct {
    Path string
}

// Cookie backend (client-side, signed)
type CookieSessionStore struct {
    SecretKey []byte
}
```

---

## 4. Configuration

```go
// Settings
{
    "sessions": {
        "backend": "database",  // database, cache, file, cookie
        "cookie_name": "sessionid",
        "cookie_age": 1209600,  // 2 weeks in seconds
        "cookie_secure": true,  // HTTPS only
        "cookie_httponly": true,
        "cookie_samesite": "Lax"
    }
}
```

---

## 5. Middleware

```go
// SessionMiddleware loads/saves sessions
func SessionMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        session := sessions.Load(r)
        ctx := context.WithValue(r.Context(), sessionKey, session)
        r = r.WithContext(ctx)
        
        next.ServeHTTP(w, r)
        
        session.Save()
        // Set cookie if needed
    })
}
```

---

## 6. Django Comparison

**Django:**
```python
# Get/set session
request.session['cart_id'] = cart_id
cart_id = request.session.get('cart_id')

# Delete
del request.session['cart_id']

# Clear
request.session.flush()
```

**Django-DRF-Go:**
```go
session := sessions.Get(r)
session.Set("cart_id", cartID)
cartID := session.Get("cart_id")

session.Delete("cart_id")
session.Clear()
```

---

## 7. Session Cleanup

```bash
# Clear expired sessions
./manage clearsessions
```

---

## 8. Related PRDs

- [03_PRD_Core_Middleware.md](./03_PRD_Core_Middleware.md) - Session middleware
- [24_PRD_Contrib_Auth.md](./24_PRD_Contrib_Auth.md) - Auth sessions
