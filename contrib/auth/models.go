package auth

import (
	"reflect"
	"strings"
	"time"

	"github.com/anuragcarret/djang-drf-go/orm/models"
	"github.com/anuragcarret/djang-drf-go/orm/queryset"
)

// Authenticatable is the interface for models that can be used for authentication
type Authenticatable interface {
	queryset.ModelInterface
	CheckPassword(password string) bool
	SetPassword(password string)
	GetID() uint64
	GetIsActive() bool
	GetIsStaff() bool
	GetIsSuperuser() bool
}

func init() {
	models.RegisterModel("auth", &User{})
	models.RegisterModel("auth", &Group{})
	models.RegisterModel("auth", &Permission{})
	models.RegisterModel("auth", &OutstandingToken{})
	models.RegisterModel("auth", &BlacklistedToken{})
}

type OutstandingToken struct {
	models.Model
	UserID uint64 `drf:"user_id"`
	JTI    string `drf:"jti;unique"`
	Token  string `drf:"token"`
	Exp    int64  `drf:"exp"`
}

func (o *OutstandingToken) TableName() string { return "go_outstanding_tokens" }

type BlacklistedToken struct {
	models.Model
	TokenID uint64 `drf:"token_id;unique"`
	Token   string `drf:"token"`
}

func (b *BlacklistedToken) TableName() string { return "go_blacklisted_tokens" }

// User is the default user model
type User struct {
	models.Model

	Username    string     `drf:"username;max_length=150;unique"`
	Email       string     `drf:"email;max_length=254;unique"`
	Password    string     `drf:"password;max_length=128;write_only"`
	FirstName   string     `drf:"first_name;max_length=30;blank"`
	LastName    string     `drf:"last_name;max_length=150;blank"`
	IsActive    bool       `drf:"is_active;default=true"`
	IsStaff     bool       `drf:"is_staff;default=false"`
	IsSuperuser bool       `drf:"is_superuser;default=false"`
	LastLogin   *time.Time `drf:"last_login;null"`
	DateJoined  time.Time  `drf:"date_joined;auto_now_add"`
}

func (u *User) TableName() string { return "go_users" }

func (u *User) SetPassword(raw string) {
	u.Password = MakePassword(raw)
}

func (u *User) CheckPassword(raw string) bool {
	return CheckPassword(raw, u.Password)
}

func (u *User) GetIsActive() bool    { return u.IsActive }
func (u *User) GetIsStaff() bool     { return u.IsStaff }
func (u *User) GetIsSuperuser() bool { return u.IsSuperuser }
func (u *User) GetID() uint64        { return u.ID }

// FindFieldByColumn is a helper to find a struct field by its DB column name
func FindFieldByColumn(v reflect.Value, col string) (reflect.Value, bool) {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("drf")
		if tag != "" {
			parts := strings.Split(tag, ";")
			if parts[0] == col {
				return v.Field(i), true
			}
		}
		if field.Name == col {
			return v.Field(i), true
		}

		// Recurse into embedded structs
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			if sub, ok := FindFieldByColumn(v.Field(i), col); ok {
				return sub, true
			}
		}
	}
	return reflect.Value{}, false
}

// Permission defines an action on a model
type Permission struct {
	models.Model
	Name     string `drf:"max_length=255"`
	Codename string `drf:"max_length=100;unique"`
}

func (p *Permission) TableName() string { return "go_permissions" }

// Group is a collection of permissions
type Group struct {
	models.Model
	Name string `drf:"max_length=150;unique"`
}

func (g *Group) TableName() string { return "go_groups" }
