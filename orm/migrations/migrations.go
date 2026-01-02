package migrations

import (
	"fmt"
	"strings"

	"github.com/anuragcarret/djang-drf-go/orm/db"
)

// Operation defines a single change to the database schema
type Operation interface {
	Apply(database *db.DB) error
	Describe() string
}

// Migration represents a set of operations to be applied to the database
type Migration struct {
	ID           string
	Dependencies []string
	Operations   []Operation
}

// CreateTable operation
type CreateTable struct {
	Name   string
	Fields map[string]string
}

func (o *CreateTable) Apply(database *db.DB) error {
	return database.CreateTable(o.Name, o.Fields)
}

func (o *CreateTable) Describe() string {
	return "Create table " + o.Name
}

// AddField operation
type AddField struct {
	TableName string
	FieldName string
	FieldType string
}

func (o *AddField) Apply(database *db.DB) error {
	query := "ALTER TABLE " + o.TableName + " ADD COLUMN " + o.FieldName + " " + o.FieldType
	_, err := database.Exec(query)
	return err
}

func (o *AddField) Describe() string {
	return "Add field " + o.FieldName + " to " + o.TableName
}

// AlterField operation
type AlterField struct {
	TableName string
	FieldName string
	FieldType string
}

func (o *AlterField) Apply(database *db.DB) error {
	// Parse FieldType: e.g. "TEXT UNIQUE NOT NULL"
	parts := strings.Split(o.FieldType, " ")
	baseType := parts[0]
	isUnique := false
	isNotNull := false

	for _, p := range parts {
		if p == "UNIQUE" {
			isUnique = true
		}
		if p == "NOT" { // NOT NULL
			isNotNull = true
		}
	}
	if strings.Contains(o.FieldType, "TIMESTAMP WITH TIME ZONE") {
		baseType = "TIMESTAMP WITH TIME ZONE"
	}

	var query string

	// 1. Alter Type
	if !strings.Contains(o.FieldType, "SERIAL") && !strings.Contains(o.FieldType, "PRIMARY KEY") {
		query = fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s USING %s::%s", o.TableName, o.FieldName, baseType, o.FieldName, baseType)
		if _, err := database.Exec(query); err != nil {
			return err
		}
	}

	// 2. Alter Not Null
	if !strings.Contains(o.FieldType, "PRIMARY KEY") {
		if isNotNull {
			query = fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET NOT NULL", o.TableName, o.FieldName)
		} else {
			query = fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL", o.TableName, o.FieldName)
		}
		if _, err := database.Exec(query); err != nil {
			return err
		}
	}

	// 3. Alter Unique (Simplification: always try to add if requested, might fail if exists)
	if isUnique {
		// Postgres requires adding a constraint for UNIQUE
		constraintName := fmt.Sprintf("uni_%s_%s", o.TableName, o.FieldName)
		database.Exec(fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s", o.TableName, constraintName))
		query = fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s UNIQUE (%s)", o.TableName, constraintName, o.FieldName)
		if _, err := database.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func (o *AlterField) Describe() string {
	return "Alter field " + o.FieldName + " on " + o.TableName
}

// RemoveField operation
type RemoveField struct {
	TableName string
	FieldName string
}

func (o *RemoveField) Apply(database *db.DB) error {
	query := fmt.Sprintf("ALTER TABLE %s DROP COLUMN IF EXISTS %s CASCADE", o.TableName, o.FieldName)
	_, err := database.Exec(query)
	return err
}

func (o *RemoveField) Describe() string {
	return "Remove field " + o.FieldName + " from " + o.TableName
}

// RunSQL operation
type RunSQL struct {
	SQL string
}

func (o *RunSQL) Apply(database *db.DB) error {
	_, err := database.Exec(o.SQL)
	return err
}

func (o *RunSQL) Describe() string {
	return "Run custom SQL"
}

// Registry stores all migrations across apps
type Registry struct {
	migrations map[string][]*Migration
}

func NewRegistry() *Registry {
	return &Registry{
		migrations: make(map[string][]*Migration),
	}
}

var GlobalRegistry = NewRegistry()

func (r *Registry) Register(appLabel string, m *Migration) {
	r.migrations[appLabel] = append(r.migrations[appLabel], m)
}

func (r *Registry) All() []*Migration {
	var all []*Migration
	for _, ms := range r.migrations {
		all = append(all, ms...)
	}
	return all
}
