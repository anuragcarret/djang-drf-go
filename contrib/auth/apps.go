package auth

import (
	"github.com/anuragcarret/djang-drf-go/core/apps"
)

type AuthApp struct{}

func (a *AuthApp) AppConfig() *apps.AppConfig {
	return &apps.AppConfig{
		Name:  "auth",
		Label: "auth",
	}
}

func (a *AuthApp) Ready() error {
	return nil
}
