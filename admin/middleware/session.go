package middleware

import (
	"context"
	"net/http"

	"github.com/anuragcarret/djang-drf-go/admin/sessions"
)

// SessionMiddleware loads session data from cookies
func SessionMiddleware(store sessions.SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Load session from cookie
			cookie, err := r.Cookie("admin_session")
			if err == nil && cookie.Value != "" {
				session, err := store.Get(cookie.Value)
				if err == nil && session != nil {
					// Add session to context
					ctx := context.WithValue(r.Context(), "admin_session", session)
					ctx = context.WithValue(ctx, "session_id", cookie.Value)
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
