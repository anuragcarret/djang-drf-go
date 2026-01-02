package auth

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/anuragcarret/djang-drf-go/orm/db"
)

type contextKey string

const UserContextKey contextKey = "user"

// AuthenticationMiddleware handles JWT verification and injects the user into the request context
func AuthenticationMiddleware(database *db.DB, userModel interface{}) func(http.Handler) http.Handler {
	modelType := reflect.TypeOf(userModel)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	return func(next http.Handler) http.Handler {
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

			// Fetch user from DB
			instance := reflect.New(modelType).Interface().(Authenticatable)
			tableName := "go_users"
			if t, ok := instance.(interface{ TableName() string }); ok {
				tableName = t.TableName()
			}

			// Query user
			sql := fmt.Sprintf("SELECT * FROM %s WHERE id = $1 LIMIT 1", tableName)
			rows, err := database.Query(sql, claims.UserID)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			defer rows.Close()

			if rows.Next() {
				// We need a helper to scan into dynamic model
				// Since we don't have a generic one, use a simplified version for now
				// or just use reflection to find fields.

				// Re-using the logic from auth_views.go's authenticateUser (simplified)
				val := reflect.ValueOf(instance).Elem()
				cols, _ := rows.Columns()
				dest := make([]interface{}, len(cols))

				for i, col := range cols {
					fieldVal, ok := FindFieldByColumn(val, col)
					if ok && fieldVal.CanAddr() {
						dest[i] = fieldVal.Addr().Interface()
					} else {
						var ignored interface{}
						dest[i] = &ignored
					}
				}

				if err := rows.Scan(dest...); err == nil {
					ctx := context.WithValue(r.Context(), UserContextKey, instance)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
