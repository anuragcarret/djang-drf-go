package urls

import (
	"fmt"
	"strconv"
)

// Converter defines how to match and parse URL parameters
type Converter interface {
	// Regex returns the pattern to match
	Regex() string

	// ToGo converts string to Go type
	ToGo(value string) (interface{}, error)

	// ToURL converts Go type to URL string
	ToURL(value interface{}) (string, error)
}

// Built-in converters
var converters = map[string]Converter{
	"int":  IntConverter{},
	"str":  StringConverter{},
	"slug": SlugConverter{},
	"uuid": UUIDConverter{},
	"path": PathConverter{},
}

// RegisterConverter adds a new path converter
func RegisterConverter(name string, c Converter) {
	converters[name] = c
}

// GetConverter retrieves a converter by name
func GetConverter(name string) (Converter, bool) {
	c, ok := converters[name]
	return c, ok
}

// IntConverter matches positive integers
type IntConverter struct{}

func (c IntConverter) Regex() string { return `[0-9]+` }
func (c IntConverter) ToGo(value string) (interface{}, error) {
	return strconv.Atoi(value)
}
func (c IntConverter) ToURL(value interface{}) (string, error) {
	return fmt.Sprintf("%d", value), nil
}

// StringConverter matches any string without slashes
type StringConverter struct{}

func (c StringConverter) Regex() string                          { return `[^/]+` }
func (c StringConverter) ToGo(value string) (interface{}, error) { return value, nil }
func (c StringConverter) ToURL(value interface{}) (string, error) {
	return fmt.Sprintf("%v", value), nil
}

// SlugConverter matches slug-formatted strings
type SlugConverter struct{}

func (c SlugConverter) Regex() string                           { return `[-a-zA-Z0-9_]+` }
func (c SlugConverter) ToGo(value string) (interface{}, error)  { return value, nil }
func (c SlugConverter) ToURL(value interface{}) (string, error) { return fmt.Sprintf("%v", value), nil }

// UUIDConverter matches UUIDs
type UUIDConverter struct{}

func (c UUIDConverter) Regex() string {
	return `[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`
}
func (c UUIDConverter) ToGo(value string) (interface{}, error)  { return value, nil }
func (c UUIDConverter) ToURL(value interface{}) (string, error) { return fmt.Sprintf("%v", value), nil }

// PathConverter matches any string including slashes
type PathConverter struct{}

func (c PathConverter) Regex() string                           { return `.+` }
func (c PathConverter) ToGo(value string) (interface{}, error)  { return value, nil }
func (c PathConverter) ToURL(value interface{}) (string, error) { return fmt.Sprintf("%v", value), nil }
