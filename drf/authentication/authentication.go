package authentication

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"
)

var (
	ErrInvalidToken       = errors.New("invalid authentication token")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrMissingAuth        = errors.New("authentication credentials not provided")
)

// Authenticator is the interface for all authentication backends
type Authenticator interface {
	Authenticate(r *http.Request) (interface{}, error)
}

// TokenStore defines the interface for token storage/validation
type TokenStore interface {
	ValidateToken(token string) (user interface{}, err error)
	CreateToken(user interface{}) (string, error)
	RevokeToken(token string) error
}

// SessionStore defines the interface for session storage
type SessionStore interface {
	GetSession(sessionID string) (user interface{}, err error)
	CreateSession(user interface{}) (sessionID string, err error)
	DestroySession(sessionID string) error
}

// UserStore defines the interface for user lookup
type UserStore interface {
	GetUserByCredentials(username, password string) (interface{}, error)
	GetUserByID(id interface{}) (interface{}, error)
}

// TokenAuthentication authenticates users via Authorization header with tokens
type TokenAuthentication struct {
	TokenStore TokenStore
	Keyword    string // "Token" or "Bearer"
}

func NewTokenAuthentication(store TokenStore) *TokenAuthentication {
	return &TokenAuthentication{
		TokenStore: store,
		Keyword:    "Token", // DRF default is "Token"
	}
}

func (a *TokenAuthentication) Authenticate(r *http.Request) (interface{}, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, ErrMissingAuth
	}

	// Parse "Token <token>" or "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return nil, ErrInvalidToken
	}

	keyword := parts[0]
	token := parts[1]

	// Accept both "Token" and "Bearer" keywords
	if keyword != a.Keyword && keyword != "Bearer" && keyword != "Token" {
		return nil, ErrInvalidToken
	}

	user, err := a.TokenStore.ValidateToken(token)
	if err != nil {
		return nil, ErrInvalidToken
	}

	return user, nil
}

// SessionAuthentication authenticates via session cookies
type SessionAuthentication struct {
	SessionStore  SessionStore
	CookieName    string
	CSRFProtected bool
}

func NewSessionAuthentication(store SessionStore) *SessionAuthentication {
	return &SessionAuthentication{
		SessionStore:  store,
		CookieName:    "sessionid",
		CSRFProtected: true,
	}
}

func (a *SessionAuthentication) Authenticate(r *http.Request) (interface{}, error) {
	cookie, err := r.Cookie(a.CookieName)
	if err != nil {
		return nil, ErrMissingAuth
	}

	sessionID := cookie.Value
	if sessionID == "" {
		return nil, ErrInvalidToken
	}

	user, err := a.SessionStore.GetSession(sessionID)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// TODO: Add CSRF validation if CSRFProtected is true
	// This would check the X-CSRFToken header against the session

	return user, nil
}

// BasicAuthentication authenticates via HTTP Basic Auth
type BasicAuthentication struct {
	UserStore UserStore
}

func NewBasicAuthentication(store UserStore) *BasicAuthentication {
	return &BasicAuthentication{
		UserStore: store,
	}
}

func (a *BasicAuthentication) Authenticate(r *http.Request) (interface{}, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, ErrMissingAuth
	}

	// Parse "Basic <base64>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Basic" {
		return nil, ErrInvalidCredentials
	}

	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Parse "username:password"
	credentials := strings.SplitN(string(decoded), ":", 2)
	if len(credentials) != 2 {
		return nil, ErrInvalidCredentials
	}

	username := credentials[0]
	password := credentials[1]

	user, err := a.UserStore.GetUserByCredentials(username, password)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// JWTAuthentication authenticates via JSON Web Tokens
type JWTAuthentication struct {
	SecretKey     string
	Algorithm     string // HS256, RS256, etc.
	UserStore     UserStore
	TokenLifetime time.Duration
}

func NewJWTAuthentication(secretKey string, userStore UserStore) *JWTAuthentication {
	return &JWTAuthentication{
		SecretKey:     secretKey,
		Algorithm:     "HS256",
		UserStore:     userStore,
		TokenLifetime: 24 * time.Hour,
	}
}

func (a *JWTAuthentication) Authenticate(r *http.Request) (interface{}, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, ErrMissingAuth
	}

	// Parse "Bearer <jwt>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, ErrInvalidToken
	}

	tokenString := parts[1]

	// TODO: Implement JWT parsing and validation
	// This would use a library like golang-jwt/jwt
	// For now, return placeholder
	_ = tokenString
	return nil, errors.New("JWT authentication not fully implemented yet")
}

// AuthenticationMiddleware tries multiple authentication backends
func AuthenticationMiddleware(authenticators []Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var user interface{}

			// Try each authenticator until one succeeds
			for _, auth := range authenticators {
				u, err := auth.Authenticate(r)
				if err == nil && u != nil {
					user = u
					break
				}
				// Continue to next authenticator if this one fails
			}

			// Set user in request context (even if nil for anonymous)
			ctx := context.WithValue(r.Context(), "user", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFrom Request retrieves the authenticated user from request context
func GetUserFromRequest(r *http.Request) interface{} {
	return r.Context().Value("user")
}

// IsAuthenticated checks if request has an authenticated user
func IsAuthenticated(r *http.Request) bool {
	user := GetUserFromRequest(r)
	return user != nil
}
