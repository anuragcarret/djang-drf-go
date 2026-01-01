package views

import (
	"encoding/json"

	"github.com/anuragcarret/djang-drf-go/contrib/auth"
	"github.com/anuragcarret/djang-drf-go/drf/views"
	"github.com/anuragcarret/djang-drf-go/orm/db"
	"github.com/anuragcarret/djang-drf-go/orm/queryset"
)

type TokenObtainPairView struct {
	views.BaseAPIView
	DB *db.DB
}

type TokenObtainRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (v *TokenObtainPairView) Post(c *views.Context) views.Response {
	var req TokenObtainRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		return views.BadRequest("Invalid request body")
	}

	qs := queryset.NewQuerySet[*auth.User](v.DB)
	user, err := qs.Filter(queryset.Q{"username": req.Username}).Get()
	if err != nil {
		return views.Forbidden("Invalid credentials")
	}

	if !user.CheckPassword(req.Password) {
		return views.Forbidden("Invalid credentials")
	}

	access, refresh, err := auth.GenerateTokenPair(v.DB, user.ID)
	if err != nil {
		return views.BadRequest("Failed to generate tokens")
	}

	return views.OK(map[string]string{
		"access":  access,
		"refresh": refresh,
	})
}

type TokenRefreshView struct {
	views.BaseAPIView
	DB *db.DB
}

type TokenRefreshRequest struct {
	Refresh string `json:"refresh"`
}

func (v *TokenRefreshView) Post(c *views.Context) views.Response {
	var req TokenRefreshRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		return views.BadRequest("Invalid request body")
	}

	access, err := auth.RefreshToken(v.DB, req.Refresh)
	if err != nil {
		return views.Forbidden(err.Error())
	}

	return views.OK(map[string]string{
		"access": access,
	})
}
