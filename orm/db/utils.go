package db

import (
	"fmt"
	"strings"
)

// CreateTable generates and executes a CREATE TABLE statement for a model
// This is a simplified version for the demo app.
// In a full implementation, this would be part of a migrations system.
func (db *DB) CreateTable(tableName string, fields map[string]string) error {
	var cols []string
	for name, typ := range fields {
		cols = append(cols, fmt.Sprintf("%s %s", name, typ))
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", tableName, strings.Join(cols, ", "))
	_, err := db.Exec(query)
	return err
}
