package auth

import (
	"time"

	"github.com/anuragcarret/djang-drf-go/orm/models"
)

// User is the default user model
type User struct {
	models.Model

	Username    string     `drf:"max_length=150;unique"`
	Email       string     `drf:"max_length=254;unique"`
	Password    string     `drf:"max_length=128"`
	FirstName   string     `drf:"max_length=30;blank"`
	LastName    string     `drf:"max_length=150;blank"`
	IsActive    bool       `drf:"default=true"`
	IsStaff     bool       `drf:"default=false"`
	IsSuperuser bool       `drf:"default=false"`
	LastLogin   *time.Time `drf:"null"`
	DateJoined  time.Time  `drf:"auto_now_add"`
}

func (u *User) SetPassword(raw string) {
	u.Password = MakePassword(raw)
}

func (u *User) CheckPassword(raw string) bool {
	return CheckPassword(raw, u.Password)
}
