package permissions

import (
	"net/http"
)

// Permission defines the interface for all permission classes
type Permission interface {
	HasPermission(r *http.Request, view interface{}) bool
	HasObjectPermission(r *http.Request, view interface{}, obj interface{}) bool
}

// BasePermission provides default implementations
type BasePermission struct{}

func (p *BasePermission) HasPermission(r *http.Request, view interface{}) bool {
	return true
}

func (p *BasePermission) HasObjectPermission(r *http.Request, view interface{}, obj interface{}) bool {
	return p.HasPermission(r, view)
}

// AllowAny allows unrestricted access
type AllowAny struct {
	BasePermission
}

func (p *AllowAny) HasPermission(r *http.Request, view interface{}) bool {
	return true
}

// IsAuthenticated allows only authenticated users
type IsAuthenticated struct {
	BasePermission
}

func (p *IsAuthenticated) HasPermission(r *http.Request, view interface{}) bool {
	user := r.Context().Value("user")
	return user != nil
}

// IsAuthenticatedOrReadOnly allows read-only for anonymous, full access for authenticated
type IsAuthenticatedOrReadOnly struct {
	BasePermission
}

func (p *IsAuthenticatedOrReadOnly) HasPermission(r *http.Request, view interface{}) bool {
	// Allow safe methods for everyone
	if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
		return true
	}

	// Require authentication for unsafe methods
	user := r.Context().Value("user")
	return user != nil
}

// IsAdminUser allows only staff/superuser access
type IsAdminUser struct {
	BasePermission
}

func (p *IsAdminUser) HasPermission(r *http.Request, view interface{}) bool {
	user := r.Context().Value("user")
	if user == nil {
		return false
	}

	// Check if user has admin privileges
	// This assumes the user object has IsStaff or IsSuperuser fields
	type AdminUser interface {
		GetIsStaff() bool
		GetIsSuperuser() bool
	}

	if adminUser, ok := user.(AdminUser); ok {
		return adminUser.GetIsStaff() || adminUser.GetIsSuperuser()
	}

	return false
}

// DjangoModelPermissions checks Django-style model permissions
type DjangoModelPermissions struct {
	BasePermission
	PermissionMap map[string][]string
}

func NewDjangoModelPermissions() *DjangoModelPermissions {
	return &DjangoModelPermissions{
		PermissionMap: map[string][]string{
			"GET":    {},         // No permissions required for read
			"POST":   {"add"},    // add_modelname
			"PUT":    {"change"}, // change_modelname
			"PATCH":  {"change"}, // change_modelname
			"DELETE": {"delete"}, // delete_modelname
		},
	}
}

func (p *DjangoModelPermissions) HasPermission(r *http.Request, view interface{}) bool {
	user := r.Context().Value("user")
	if user == nil {
		return false
	}

	requiredPerms, ok := p.PermissionMap[r.Method]
	if !ok || len(requiredPerms) == 0 {
		return true // No permissions required
	}

	// Check if user has required permissions
	// This assumes the user object has a HasPerm method
	type PermissionChecker interface {
		HasPerm(perm string) bool
	}

	if checker, ok := user.(PermissionChecker); ok {
		for _, perm := range requiredPerms {
			if !checker.HasPerm(perm) {
				return false
			}
		}
		return true
	}

	return false
}

// DjangoObjectPermissions extends model permissions to object level
type DjangoObjectPermissions struct {
	DjangoModelPermissions
}

func NewDjangoObjectPermissions() *DjangoObjectPermissions {
	return &DjangoObjectPermissions{
		DjangoModelPermissions: *NewDjangoModelPermissions(),
	}
}

func (p *DjangoObjectPermissions) HasObjectPermission(r *http.Request, view interface{}, obj interface{}) bool {
	user := r.Context().Value("user")
	if user == nil {
		return false
	}

	requiredPerms, ok := p.PermissionMap[r.Method]
	if !ok || len(requiredPerms) == 0 {
		return true
	}

	// Check object-level permissions
	// This assumes the user object has a HasObjectPerm method
	type ObjectPermissionChecker interface {
		HasObjectPerm(perm string, obj interface{}) bool
	}

	if checker, ok := user.(ObjectPermissionChecker); ok {
		for _, perm := range requiredPerms {
			if !checker.HasObjectPerm(perm, obj) {
				return false
			}
		}
		return true
	}

	return false
}

// IsOwnerOrReadOnly allows read access to all, write access only to owner
type IsOwnerOrReadOnly struct {
	BasePermission
}

func (p *IsOwnerOrReadOnly) HasObjectPermission(r *http.Request, view interface{}, obj interface{}) bool {
	// Allow safe methods for everyone
	if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
		return true
	}

	user := r.Context().Value("user")
	if user == nil {
		return false
	}

	// Check if obj has an owner/user field matching the current user
	type OwnedObject interface {
		GetOwner() interface{}
	}

	if owned, ok := obj.(OwnedObject); ok {
		owner := owned.GetOwner()

		// Compare user IDs
		type IDGetter interface {
			GetID() interface{}
		}

		if userWithID, ok := user.(IDGetter); ok {
			if ownerWithID, ok := owner.(IDGetter); ok {
				return userWithID.GetID() == ownerWithID.GetID()
			}
		}

		// Fallback to direct comparison
		return owner == user
	}

	return false
}

// CheckPermissions helper function to check all permissions
func CheckPermissions(r *http.Request, view interface{}, permissions []Permission) bool {
	for _, perm := range permissions {
		if !perm.HasPermission(r, view) {
			return false
		}
	}
	return true
}

// CheckObjectPermissions helper function to check all object permissions
func CheckObjectPermissions(r *http.Request, view interface{}, obj interface{}, permissions []Permission) bool {
	for _, perm := range permissions {
		if !perm.HasObjectPermission(r, view, obj) {
			return false
		}
	}
	return true
}
