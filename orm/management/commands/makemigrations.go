package management

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/anuragcarret/djang-drf-go/orm/db"
	"github.com/anuragcarret/djang-drf-go/orm/migrations"
)

type MakemigrationsCommand struct {
	database *db.DB
}

func NewMakemigrationsCommand(database *db.DB) *MakemigrationsCommand {
	return &MakemigrationsCommand{database: database}
}

func (c *MakemigrationsCommand) Name() string { return "makemigrations" }
func (c *MakemigrationsCommand) Help() string { return "Generate migration files from model changes" }

func (c *MakemigrationsCommand) Run(ctx context.Context, args []string) error {
	detector := migrations.NewAutodetector(c.database)
	ops, err := detector.Changes()
	if err != nil {
		return err
	}

	if len(ops) == 0 {
		fmt.Println("No changes detected")
		return nil
	}

	// For demo, we assume the app label is 'demo' and dir is 'migrations'
	cwd, _ := os.Getwd()
	appDir := filepath.Join(cwd, "migrations")

	writer := migrations.NewWriter("demo", appDir)
	filename, err := writer.Write(ops)
	if err != nil {
		return err
	}

	fmt.Printf("Generated migration: %s\n", filename)
	return nil
}
