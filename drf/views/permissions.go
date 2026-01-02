package views

import (
	"github.com/anuragcarret/djang-drf-go/contrib/auth"
)

// Permission defines the interface for DRF-style permissions
type Permission interface {
	HasPermission(c *Context) bool
}

// AllowAny allows any request
type AllowAny struct{}

func (p *AllowAny) HasPermission(c *Context) bool {
	return true
}

// IsAuthenticated allows only authenticated users
type IsAuthenticated struct{}

func (p *IsAuthenticated) HasPermission(c *Context) bool {
	if c.User == nil {
		return false
	}
	// Check if it's a valid authenticatable object
	user, ok := c.User.(auth.Authenticatable)
	return ok && user.GetIsActive()
}

// IsAdminUser allows only staff/admin users
type IsAdminUser struct{}

func (p *IsAdminUser) HasPermission(c *Context) bool {
	user, ok := c.User.(auth.Authenticatable)
	return ok && user.GetIsActive() && (user.GetIsStaff() || user.GetIsSuperuser())
}
