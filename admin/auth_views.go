package admin

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"reflect"
	"time"

	"github.com/anuragcarret/djang-drf-go/admin/sessions"
	"github.com/anuragcarret/djang-drf-go/contrib/auth"
	"github.com/anuragcarret/djang-drf-go/orm/db"
)

// LoginView handles admin login
type LoginView struct {
	Store     sessions.SessionStore
	DB        *db.DB
	Templates *template.Template
	UserModel reflect.Type // The type of user model to use for authentication
}

func (v *LoginView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		v.handleLogin(w, r)
		return
	}

	// Render login form
	v.renderLoginForm(w, r, "", "")
}

func (v *LoginView) handleLogin(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	remember := r.FormValue("remember") == "true"

	// Authenticate user
	user, err := v.authenticateUser(username, password)
	if err != nil {
		v.renderLoginForm(w, r, "Invalid username or password", username)
		return
	}

	// Get values from user using reflection (handles both auth.User and custom embedding)
	v_val := reflect.ValueOf(user)
	if v_val.Kind() == reflect.Ptr {
		v_val = v_val.Elem()
	}

	// Helper to get field value by name
	getField := func(name string) interface{} {
		// Use reflection to find field
		f := v_val.FieldByName(name)
		if !f.IsValid() {
			// Find in embedded if not found
			for i := 0; i < v_val.NumField(); i++ {
				field := v_val.Field(i)
				if field.Kind() == reflect.Struct {
					subF := field.FieldByName(name)
					if subF.IsValid() {
						return subF.Interface()
					}
				}
			}
			return nil
		}
		return f.Interface()
	}

	isActive, _ := getField("IsActive").(bool)
	if !isActive {
		v.renderLoginForm(w, r, "This account has been disabled", username)
		return
	}

	isStaff, _ := getField("IsStaff").(bool)
	isSuperuser, _ := getField("IsSuperuser").(bool)
	if !isStaff && !isSuperuser {
		v.renderLoginForm(w, r, "You don't have permission to access the admin panel", username)
		return
	}

	// Create session
	sessionID := generateSecureSessionID()
	csrfToken := generateCSRFToken()

	userID := user.GetID()
	username_val, _ := getField("Username").(string)

	session := &sessions.SessionData{
		UserID:      userID,
		Username:    username_val,
		IsStaff:     isStaff,
		IsSuperuser: isSuperuser,
		CreatedAt:   time.Now(),
		LastAccess:  time.Now(),
		CSRFToken:   csrfToken,
	}

	// Set expiry based on "remember me"
	expiry := 2 * time.Hour // Default: 2 hours
	maxAge := 2 * 60 * 60
	if remember {
		expiry = 14 * 24 * time.Hour // 2 weeks
		maxAge = 14 * 24 * 60 * 60
	}

	v.Store.Set(sessionID, session, expiry)

	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_session",
		Value:    sessionID,
		Path:     "/admin/",
		HttpOnly: true,
		Secure:   false, // TODO: Set to true in production (HTTPS)
		SameSite: http.SameSiteStrictMode,
		MaxAge:   maxAge,
	})

	// Redirect to next or dashboard
	next := r.URL.Query().Get("next")
	if next == "" {
		next = "/admin/"
	}
	http.Redirect(w, r, next, http.StatusSeeOther)
}

func (v *LoginView) authenticateUser(username, password string) (auth.Authenticatable, error) {
	modelType := v.UserModel
	if modelType == nil {
		modelType = reflect.TypeOf(&auth.User{})
	}

	// Get table name using reflection
	instance := reflect.New(modelType.Elem()).Interface()
	tableName := "go_users"
	if t, ok := instance.(interface{ TableName() string }); ok {
		tableName = t.TableName()
	}

	// Query the user
	sql := fmt.Sprintf("SELECT * FROM %s WHERE username = $1 LIMIT 1", tableName)
	rows, err := v.DB.Query(sql, username)
	if err != nil {
		return nil, fmt.Errorf("database query error: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("user not found")
	}

	// Create new user object
	user := reflect.New(modelType.Elem()).Interface().(auth.Authenticatable)

	// Manual scan since we don't have a generic scan in DB yet
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	val := reflect.ValueOf(user).Elem()
	dest := make([]interface{}, len(cols))

	// Prepare scan destinations
	for i, col := range cols {
		fieldVal, ok := auth.FindFieldByColumn(val, col)
		if ok && fieldVal.CanAddr() {
			dest[i] = fieldVal.Addr().Interface()
		} else {
			var ignored interface{}
			dest[i] = &ignored
		}
	}

	if err := rows.Scan(dest...); err != nil {
		return nil, fmt.Errorf("scan error: %v", err)
	}

	// Check password
	if !user.CheckPassword(password) {
		return nil, fmt.Errorf("invalid password")
	}

	return user, nil
}

func (v *LoginView) renderLoginForm(w http.ResponseWriter, r *http.Request, errorMsg, username string) {
	data := map[string]interface{}{
		"Error":    errorMsg,
		"Username": username,
	}

	if v.Templates != nil {
		v.Templates.ExecuteTemplate(w, "login.html", data)
	} else {
		// Fallback if templates not loaded properly
		http.Error(w, "Login page not available", http.StatusInternalServerError)
	}
}

// LogoutView handles admin logout
type LogoutView struct {
	Store sessions.SessionStore
}

func (v *LogoutView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get session cookie
	cookie, err := r.Cookie("admin_session")
	if err == nil {
		// Delete session from store
		v.Store.Delete(cookie.Value)
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_session",
		Value:    "",
		Path:     "/admin/",
		HttpOnly: true,
		MaxAge:   -1, // Delete cookie
	})

	// Redirect to login
	http.Redirect(w, r, "/admin/login/", http.StatusSeeOther)
}

// generateSecureSessionID creates a cryptographically secure random session ID
func generateSecureSessionID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// generateCSRFToken creates a cryptographically secure random CSRF token
func generateCSRFToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
