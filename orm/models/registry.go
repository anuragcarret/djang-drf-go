package models

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/anuragcarret/djang-drf-go/orm/fields"
)

// ModelInfo holds cached metadata about a model
type ModelInfo struct {
	Type      reflect.Type
	TableName string
	Fields    []fields.Field
	PKField   fields.Field
	Meta      *ModelMeta
}

// Registry manages model metadata
type Registry struct {
	models map[reflect.Type]*ModelInfo
	mu     sync.RWMutex
}

var globalRegistry = NewRegistry()

func NewRegistry() *Registry {
	return &Registry{
		models: make(map[reflect.Type]*ModelInfo),
	}
}

// Register adds a model to the registry and introspects its fields
func (r *Registry) Register(m ModelInterface) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	t := reflect.TypeOf(m)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if _, ok := r.models[t]; ok {
		return nil // Already registered
	}

	info := &ModelInfo{
		Type:      t,
		TableName: m.TableName(),
		Meta:      m.Meta(),
	}

	// Introspect fields
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Handle embedded Model
		if field.Anonymous && field.Type == reflect.TypeOf(Model{}) {
			r.introspectEmbedded(field.Type, info)
			continue
		}

		tag := field.Tag.Get("drf")
		opts := fields.ParseTag(tag)

		f := r.createField(field.Name, field.Type, opts)
		info.Fields = append(info.Fields, f)

		if opts.PrimaryKey {
			info.PKField = f
		}
	}

	r.models[t] = info
	return nil
}

func (r *Registry) introspectEmbedded(t reflect.Type, info *ModelInfo) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("drf")
		opts := fields.ParseTag(tag)

		f := r.createField(field.Name, field.Type, opts)
		info.Fields = append(info.Fields, f)

		if opts.PrimaryKey {
			info.PKField = f
		}
	}
}

func (r *Registry) createField(name string, t reflect.Type, opts *fields.FieldOptions) fields.Field {
	// Minimal implementation for now, using dummy field wrappers
	// In a real implementation, this would use a factory from orm/fields
	return &dummyField{
		name:    name,
		column:  r.columnName(name, opts),
		options: opts,
		goType:  t,
	}
}

func (r *Registry) columnName(fieldName string, opts *fields.FieldOptions) string {
	if opts.DBColumn != "" {
		return opts.DBColumn
	}
	if fieldName == "ID" {
		return "id"
	}

	var result strings.Builder
	for i, r := range fieldName {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

func (r *Registry) GetModelInfo(m ModelInterface) (*ModelInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t := reflect.TypeOf(m)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	info, ok := r.models[t]
	if !ok {
		return nil, fmt.Errorf("model %v not registered", t)
	}
	return info, nil
}

// dummyField is a placeholder to pass tests until orm/fields is fully integrated
type dummyField struct {
	name    string
	column  string
	options *fields.FieldOptions
	goType  reflect.Type
}

func (f *dummyField) Name() string                                        { return f.name }
func (f *dummyField) Column() string                                      { return f.column }
func (f *dummyField) SQLType(dialect string) string                       { return "TEXT" }
func (f *dummyField) GoType() reflect.Type                                { return f.goType }
func (f *dummyField) Options() *fields.FieldOptions                       { return f.options }
func (f *dummyField) ToDatabase(value interface{}) (interface{}, error)   { return value, nil }
func (f *dummyField) FromDatabase(value interface{}) (interface{}, error) { return value, nil }
func (f *dummyField) Validate(value interface{}) error                    { return nil }
