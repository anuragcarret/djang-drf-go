package migrations

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/anuragcarret/djang-drf-go/orm/db"
)

// Executor manages the execution of migrations
type Executor struct {
	db           *db.DB
	trackerTable string
}

func NewExecutor(database *db.DB) *Executor {
	return &Executor{
		db:           database,
		trackerTable: "go_migrations",
	}
}

// Setup creates the migration tracker table if it doesn't exist
func (e *Executor) Setup() error {
	return e.db.CreateTable(e.trackerTable, map[string]string{
		"id":         "SERIAL PRIMARY KEY",
		"name":       "VARCHAR(255) UNIQUE NOT NULL",
		"applied_at": "TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP",
	})
}

// Migrate applies a list of migrations in the correct order
func (e *Executor) Migrate(availableMigrations []*Migration) error {
	if err := e.Setup(); err != nil {
		return fmt.Errorf("failed to setup migration tracker: %w", err)
	}

	applied, err := e.getAppliedMigrations()
	if err != nil {
		return err
	}

	// Simple topological sort or just sequential for now
	// For this version, we assume linear migrations or user provides them sorted
	sort.Slice(availableMigrations, func(i, j int) bool {
		return availableMigrations[i].ID < availableMigrations[j].ID
	})

	for _, m := range availableMigrations {
		if _, ok := applied[m.ID]; ok {
			continue // Already applied
		}

		log.Printf("Applying migration %s...", m.ID)
		for _, op := range m.Operations {
			log.Printf("  - %s", op.Describe())
			if err := op.Apply(e.db); err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", m.ID, err)
			}
		}

		if err := e.markApplied(m.ID); err != nil {
			return err
		}
		log.Printf("Successfully applied %s", m.ID)
	}

	return nil
}

func (e *Executor) getAppliedMigrations() (map[string]time.Time, error) {
	rows, err := e.db.Query(fmt.Sprintf("SELECT name, applied_at FROM %s", e.trackerTable))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]time.Time)
	for rows.Next() {
		var name string
		var at time.Time
		if err := rows.Scan(&name, &at); err != nil {
			return nil, err
		}
		applied[name] = at
	}
	return applied, nil
}

func (e *Executor) markApplied(name string) error {
	_, err := e.db.Exec(fmt.Sprintf("INSERT INTO %s (name) VALUES ($1)", e.trackerTable), name)
	return err
}
