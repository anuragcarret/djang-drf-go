package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/anuragcarret/djang-drf-go/admin/sessions"
)

// AdminAuthMiddleware checks if user is authenticated and has staff/superuser permissions
func AdminAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for public paths
			if isPublicPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Check session
			session, ok := r.Context().Value("admin_session").(*sessions.SessionData)
			if !ok || session == nil {
				// Not logged in - redirect to login
				redirectToLogin(w, r)
				return
			}

			// Check staff/superuser status
			if !session.IsStaff && !session.IsSuperuser {
				http.Error(w, "Admin access required - you must be staff or superuser", http.StatusForbidden)
				return
			}

			// All checks passed
			next.ServeHTTP(w, r)
		})
	}
}

func isPublicPath(path string) bool {
	publicPaths := []string{
		"/login/",
		"/logout/",
		"/admin/login/",
		"/admin/logout/",
	}

	for _, p := range publicPaths {
		if path == p || strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	// For now, assume /admin/login/ as the standard path.
	// In a more advanced version, this could be configurable.
	loginURL := "/admin/login/"

	// We want to preserve the full original path for the 'next' parameter
	// If we are under /admin/, the path is already correct.
	nextPath := r.URL.Path
	if !strings.HasPrefix(nextPath, "/admin/") {
		nextPath = "/admin" + nextPath
	}

	fullLoginURL := loginURL + "?next=" + url.QueryEscape(nextPath)
	http.Redirect(w, r, fullLoginURL, http.StatusFound)
}
