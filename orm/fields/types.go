package fields

import (
	"fmt"
	"reflect"
)

// StringField handles string values
type StringField struct {
	BaseField
}

func (f *StringField) SQLType(dialect string) string {
	if f.options.Type != "" {
		return f.options.Type
	}
	if f.options.MaxLength > 0 {
		return fmt.Sprintf("VARCHAR(%d)", f.options.MaxLength)
	}
	return "TEXT"
}

func (f *StringField) ToDatabase(value interface{}) (interface{}, error)   { return value, nil }
func (f *StringField) FromDatabase(value interface{}) (interface{}, error) { return value, nil }

func (f *StringField) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", value)
	}

	if !f.options.Blank && str == "" {
		return fmt.Errorf("field cannot be blank")
	}

	if f.options.MaxLength > 0 && len(str) > f.options.MaxLength {
		return fmt.Errorf("length %d exceeds max_length %d", len(str), f.options.MaxLength)
	}

	return nil
}

// IntField handles integer values
type IntField struct {
	BaseField
}

func (f *IntField) SQLType(dialect string) string {
	return "INTEGER"
}

func (f *IntField) ToDatabase(value interface{}) (interface{}, error)   { return value, nil }
func (f *IntField) FromDatabase(value interface{}) (interface{}, error) { return value, nil }

func (f *IntField) Validate(value interface{}) error {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return nil
	default:
		return fmt.Errorf("expected integer, got %T", value)
	}
}

// BoolField handles boolean values
type BoolField struct {
	BaseField
}

func (f *BoolField) SQLType(dialect string) string {
	return "BOOLEAN"
}

func (f *BoolField) ToDatabase(value interface{}) (interface{}, error)   { return value, nil }
func (f *BoolField) FromDatabase(value interface{}) (interface{}, error) { return value, nil }

func (f *BoolField) Validate(value interface{}) error {
	if _, ok := value.(bool); !ok {
		return fmt.Errorf("expected boolean, got %T", value)
	}
	return nil
}
