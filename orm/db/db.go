package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// DB manages the database connection and execution
type DB struct {
	conn    *sql.DB
	dialect string
}

// NewDB creates a new database instance
func NewDB(dialect, dsn string) (*DB, error) {
	conn, err := sql.Open(dialect, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{
		conn:    conn,
		dialect: dialect,
	}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// Query executes a query and returns rows
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.conn.Query(query, args...)
}

// Exec executes a command without returning rows
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.conn.Exec(query, args...)
}

// QueryRow executes a query that returns a single row
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.conn.QueryRow(query, args...)
}

// Dialect returns the database type (e.g., "postgres")
func (db *DB) Dialect() string {
	return db.dialect
}
