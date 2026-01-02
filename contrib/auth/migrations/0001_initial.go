package migrations

import (
	"github.com/anuragcarret/djang-drf-go/orm/migrations"
)

func init() {
	migrations.GlobalRegistry.Register("auth", &migrations.Migration{
		ID: "0001_initial",
		Operations: []migrations.Operation{
			&migrations.CreateTable{
				Name: "go_users",
				Fields: map[string]string{
					"id":           "SERIAL PRIMARY KEY",
					"username":     "VARCHAR(150) UNIQUE NOT NULL",
					"email":        "VARCHAR(254) UNIQUE NOT NULL",
					"password":     "VARCHAR(128) NOT NULL",
					"first_name":   "VARCHAR(30)",
					"last_name":    "VARCHAR(150)",
					"is_active":    "BOOLEAN DEFAULT TRUE",
					"is_staff":     "BOOLEAN DEFAULT FALSE",
					"is_superuser": "BOOLEAN DEFAULT FALSE",
					"last_login":   "TIMESTAMP WITH TIME ZONE",
					"date_joined":  "TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP",
				},
			},
			&migrations.CreateTable{
				Name: "go_groups",
				Fields: map[string]string{
					"id":   "SERIAL PRIMARY KEY",
					"name": "VARCHAR(150) UNIQUE NOT NULL",
				},
			},
			&migrations.CreateTable{
				Name: "go_permissions",
				Fields: map[string]string{
					"id":       "SERIAL PRIMARY KEY",
					"name":     "VARCHAR(255) NOT NULL",
					"codename": "VARCHAR(100) UNIQUE NOT NULL",
				},
			},
		},
	})
}
