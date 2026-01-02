package urls

import (
	"github.com/anuragcarret/djang-drf-go/contrib/auth"
	"github.com/anuragcarret/djang-drf-go/contrib/auth/views"
	"github.com/anuragcarret/djang-drf-go/core/urls"
	drf_views "github.com/anuragcarret/djang-drf-go/drf/views"
	"github.com/anuragcarret/djang-drf-go/orm/db"
)

// RegisterRoutes registers the URL patterns for the auth app with a custom user model
func RegisterRoutes[T auth.Authenticatable](database *db.DB) *urls.Router {
	r := urls.NewRouter()

	tokenView := &views.TokenObtainPairView[T]{DB: database}
	refreshView := &views.TokenRefreshView{DB: database}

	r.Post("/token/", drf_views.Handler(tokenView), "token_obtain_pair")
	r.Post("/token/refresh/", drf_views.Handler(refreshView), "token_refresh")

	return r
}
