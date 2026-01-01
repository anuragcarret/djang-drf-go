package management

import (
	"context"
	"fmt"

	"github.com/anuragcarret/djang-drf-go/orm/db"
	"github.com/anuragcarret/djang-drf-go/orm/migrations"
)

type MigrateCommand struct {
	database *db.DB
}

func NewMigrateCommand(database *db.DB) *MigrateCommand {
	return &MigrateCommand{database: database}
}

func (c *MigrateCommand) Name() string { return "migrate" }
func (c *MigrateCommand) Help() string { return "Apply database migrations" }

func (c *MigrateCommand) Run(ctx context.Context, args []string) error {
	executor := migrations.NewExecutor(c.database)
	allMigrations := migrations.GlobalRegistry.All()

	fmt.Printf("Operations to perform:\n")
	fmt.Printf("  Apply all migrations: %d migrations discovered\n", len(allMigrations))

	if err := executor.Migrate(allMigrations); err != nil {
		return err
	}

	fmt.Println("Migrations applied successfully")
	return nil
}

func init() {
	// Note: In a real app, we'd need a way to pass the DB instance here.
	// For now, we'll let the user register it manually or use a global.
}
