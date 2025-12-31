package drf

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Serializer is a basic interface for data transformation.
type Serializer interface {
	ToJSON(v interface{}) ([]byte, error)
	FromJSON(data []byte, v interface{}) error
	Validate(v interface{}) error
}

// BaseSerializer provides generic serialization logic.
type BaseSerializer struct{}

// ToJSON converts a struct to JSON, respecting `json` tags.
func (s *BaseSerializer) ToJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// FromJSON converts JSON to a struct.
func (s *BaseSerializer) FromJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// Validate uses reflection to check for 'required' or other tags.
func (s *BaseSerializer) Validate(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("serialize target must be a struct")
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		tag := field.Tag.Get("serializer")
		if tag == "required" {
			fVal := val.Field(i)
			if fVal.IsZero() {
				return fmt.Errorf("field %s is required", field.Name)
			}
		}
	}
	return nil
}
