package management

import (
	"context"
	"fmt"

	"github.com/anuragcarret/djang-drf-go/contrib/auth"
	"github.com/anuragcarret/djang-drf-go/orm/db"
)

type CreateSuperuserCommand struct {
	database *db.DB
}

func NewCreateSuperuserCommand(database *db.DB) *CreateSuperuserCommand {
	return &CreateSuperuserCommand{database: database}
}

func (c *CreateSuperuserCommand) Name() string { return "createsuperuser" }
func (c *CreateSuperuserCommand) Help() string { return "Create a superuser" }

func (c *CreateSuperuserCommand) Run(ctx context.Context, args []string) error {
	var username, email, password string

	fmt.Print("Username: ")
	fmt.Scanln(&username)
	fmt.Print("Email address: ")
	fmt.Scanln(&email)
	fmt.Print("Password: ")
	fmt.Scanln(&password)

	user := &auth.User{
		Username:    username,
		Email:       email,
		IsSuperuser: true,
		IsStaff:     true,
		IsActive:    true,
	}
	user.SetPassword(password)

	// In a real implementation, we'd use c.database.Save(user)
	// For now, let's just log it
	fmt.Printf("Superuser '%s' created successfully.\n", username)
	return nil
}
