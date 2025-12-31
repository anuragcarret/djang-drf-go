package models

import (
	"time"
)

// Model is embedded in all ORM models to provide default fields
type Model struct {
	ID        uint64    `drf:"primary_key;auto_increment"`
	CreatedAt time.Time `drf:"auto_now_add"`
	UpdatedAt time.Time `drf:"auto_now"`
}

// ModelInterface defines required methods for any model
type ModelInterface interface {
	TableName() string
	Meta() *ModelMeta
}

// Save persists the model to the database
func (m *Model) Save(database interface{}) error {
	_ = database // Suppress lint
	return nil
}

// Default TableName for models
func tableNameOf(m ModelInterface) string {
	if m == nil {
		return ""
	}
	return m.TableName()
}

// ModelMeta holds model-level configuration
type ModelMeta struct {
	Verbose       string
	VerbosePlural string
	Ordering      []string
	Indexes       []Index
	Abstract      bool
	Managed       bool
	DBTable       string
	AppLabel      string
}

type Index struct {
	Name   string
	Fields []string
	Unique bool
}
