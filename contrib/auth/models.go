package auth

import (
	"time"

	"github.com/anuragcarret/djang-drf-go/orm/models"
)

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

	Username    string     `drf:"max_length=150;unique"`
	Email       string     `drf:"max_length=254;unique"`
	Password    string     `drf:"max_length=128;write_only"`
	FirstName   string     `drf:"max_length=30;blank"`
	LastName    string     `drf:"max_length=150;blank"`
	IsActive    bool       `drf:"default=true"`
	IsStaff     bool       `drf:"default=false"`
	IsSuperuser bool       `drf:"default=false"`
	LastLogin   *time.Time `drf:"null"`
	DateJoined  time.Time  `drf:"auto_now_add"`
}

func (u *User) TableName() string { return "go_users" }

func (u *User) SetPassword(raw string) {
	u.Password = MakePassword(raw)
}

func (u *User) CheckPassword(raw string) bool {
	return CheckPassword(raw, u.Password)
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
