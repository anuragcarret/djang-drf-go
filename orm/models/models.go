package models

import (
	"time"

	"github.com/anuragcarret/djang-drf-go/core/apps"
	"github.com/anuragcarret/djang-drf-go/orm/signals"
)

// Model is embedded in all ORM models to provide default fields
type Model struct {
	ID        uint64    `drf:"id;primary_key;auto_increment"`
	CreatedAt time.Time `drf:"created_at;auto_now_add"`
	UpdatedAt time.Time `drf:"updated_at;auto_now"`
}

func (m *Model) GetID() uint64 {
	return m.ID
}

// ModelInterface defines required methods for any model
type ModelInterface interface {
	TableName() string
	Meta() *ModelMeta
}

// GlobalRegistry proxy for backward compatibility
var GlobalRegistry = struct {
	Models map[string]interface{}
}{
	Models: make(map[string]interface{}),
}

func RegisterModel(appLabel string, model interface{}) {
	m, ok := model.(ModelInterface)
	if !ok {
		return
	}
	tableName := m.TableName()
	apps.Apps.RegisterModel(appLabel, model, tableName)
	GlobalRegistry.Models[tableName] = model
}

// Save persists the model to the database and triggers signals
func (m *Model) Save(database interface{}, instance interface{}) error {
	// 1. Pre-save Signal
	signals.Send(signals.PreSave, instance, instance, nil)

	// ... (Persistence logic would go here) ...
	// For now, let's assume it's successful

	// 2. Post-save Signal
	signals.Send(signals.PostSave, instance, instance, map[string]interface{}{
		"created": m.ID == 0, // Simplified creation detection
	})

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
