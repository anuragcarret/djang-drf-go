package auth

import (
	"context"
	"net/http"
	"strings"
)

const UserContextKey = "user"

// AuthenticationMiddleware handles JWT verification and injects the user into the request context
func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			next.ServeHTTP(w, r)
			return
		}

		token := parts[1]
		claims, err := ValidateToken(token)
		if err != nil || claims.Type != "access" {
			next.ServeHTTP(w, r)
			return
		}

		// TODO: Check blacklist in DB
		// SELECT 1 FROM go_blacklisted_tokens WHERE token = ?

		ctx := context.WithValue(r.Context(), UserContextKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
