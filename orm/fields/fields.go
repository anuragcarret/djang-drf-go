package fields

import (
	"reflect"
)

// Field represents a database column definition
type Field interface {
	Name() string
	Column() string
	SQLType(dialect string) string
	GoType() reflect.Type
	Options() *FieldOptions
	ToDatabase(value interface{}) (interface{}, error)
	FromDatabase(value interface{}) (interface{}, error)
	Validate(value interface{}) error
}

// FieldOptions holds configuration for a field
type FieldOptions struct {
	PrimaryKey bool
	Unique     bool
	Null       bool
	Blank      bool
	Index      bool
	Default    interface{}
	MaxLength  int
	Decimal    string
	Type       string
	Choices    []string
	Validators []string
	AutoNow    bool
	AutoNowAdd bool
	DBColumn   string
	Editable   bool
	HelpText   string
}

// BaseField provides common functionality for all fields
type BaseField struct {
	name    string
	options *FieldOptions
	goType  reflect.Type
}

func (f *BaseField) Name() string { return f.name }
func (f *BaseField) Column() string {
	if f.options.DBColumn != "" {
		return f.options.DBColumn
	}
	return f.name
}
func (f *BaseField) GoType() reflect.Type   { return f.goType }
func (f *BaseField) Options() *FieldOptions { return f.options }
