package migrations

import (
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
