package serializers

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// BaseSerializer provides the core logic for validation and serialization
type BaseSerializer struct {
	model  interface{}
	errors map[string]string
	depth  int
}

// NewSerializer creates a new serializer for a given model/struct
func NewSerializer(model interface{}) *BaseSerializer {
	return &BaseSerializer{
		model:  model,
		errors: make(map[string]string),
		depth:  0, // Default to flat serialization
	}
}

// SetDepth sets the recursion depth for nested serialization
func (s *BaseSerializer) SetDepth(depth int) {
	s.depth = depth
}

// IsValid validates the input data against the model's tags
func (s *BaseSerializer) IsValid(data map[string]interface{}) bool {
	s.errors = make(map[string]string)
	val := reflect.ValueOf(s.model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	t := val.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		drfTag := field.Tag.Get("drf")
		name := field.Name

		// Determine key name (simplified for now)
		if drfTag != "" {
			parts := strings.Split(drfTag, ";")
			if parts[0] != "" && !isOption(parts[0]) {
				name = parts[0]
			}
		}

		value, exists := data[name]

		// Check Required (unless it's a primary_key or has default)
		if !exists {
			if !hasOption(drfTag, "null") && !hasOption(drfTag, "blank") && !hasOption(drfTag, "default") {
				s.errors[name] = "This field is required."
				continue
			}
			continue
		}

		// Max Length
		if maxLenStr := getOptionValue(drfTag, "max_length"); maxLenStr != "" {
			maxLen, _ := strconv.Atoi(maxLenStr)
			if strVal, ok := value.(string); ok && len(strVal) > maxLen {
				s.errors[name] = fmt.Sprintf("Ensure this field has at most %d characters.", maxLen)
			}
		}
	}

	return len(s.errors) == 0
}

// Errors returns the validation errors
func (s *BaseSerializer) Errors() map[string]string {
	return s.errors
}

// ValidatedData returns the data that passed validation
func (s *BaseSerializer) ValidatedData() map[string]interface{} {
	// For now returns all data, but in real implementation should only return validated fields
	return nil
}

// Save creates or updates the model instance
func (s *BaseSerializer) Save() (interface{}, error) {
	if len(s.errors) > 0 {
		return nil, fmt.Errorf("cannot save invalid data")
	}

	val := reflect.ValueOf(s.model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Logic to populate model from validated data would go here
	// and calling c.DB.Save(s.model)

	return s.model, nil
}

// Data returns the serialized representation of the model
func (s *BaseSerializer) Data() interface{} {
	return Serialize(s.model, s.depth)
}

// Serialize converts a struct or slice of structs to a map/slice of maps,
// excluding fields tagged with 'write_only' in the 'drf' tag.
// It supports nested serialization up to the specified depth.
func Serialize(obj interface{}, depth int) interface{} {
	if obj == nil {
		return nil
	}

	val := reflect.ValueOf(obj)

	// Handle pointers
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Struct:
		return serializeStruct(val, depth)
	case reflect.Slice:
		res := make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i++ {
			res[i] = Serialize(val.Index(i).Interface(), depth)
		}
		return res
	case reflect.Map:
		return obj
	default:
		return obj
	}
}

func serializeStruct(v reflect.Value, depth int) map[string]interface{} {
	res := make(map[string]interface{})
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !field.IsExported() {
			continue
		}

		drfTag := field.Tag.Get("drf")
		jsonTag := field.Tag.Get("json")

		if hasOption(drfTag, "write_only") {
			continue
		}

		if field.Anonymous {
			embedded := serializeStruct(v.Field(i), depth)
			for k, val := range embedded {
				res[k] = val
			}
			continue
		}

		name := field.Name
		if jsonTag != "" && jsonTag != "-" {
			name = strings.Split(jsonTag, ",")[0]
		}

		if drfTag != "" {
			parts := strings.Split(drfTag, ";")
			if parts[0] != "" && !isOption(parts[0]) {
				name = parts[0]
			}
		}

		if jsonTag == "-" {
			continue
		}

		fieldVal := v.Field(i)

		// Handle Relational Fields
		isRel := hasOption(drfTag, "relation") || hasOption(drfTag, "m2m")
		if isRel {
			// Only serialize nested if depth > 0 and field is not zero/empty
			if depth > 0 && !isZero(fieldVal) {
				res[name] = Serialize(fieldVal.Interface(), depth-1)
			}
			// If depth is 0, we skip these rich relations to avoid heavy objects
			continue
		}

		res[name] = fieldVal.Interface()
	}
	return res
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Slice, reflect.Map:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return v.IsZero()
	}
}

func hasOption(tag, option string) bool {
	parts := strings.Split(tag, ";")
	for _, p := range parts {
		if p == option || strings.HasPrefix(p, option+"=") {
			return true
		}
	}
	return false
}

func getOptionValue(tag, option string) string {
	parts := strings.Split(tag, ";")
	for _, p := range parts {
		if strings.HasPrefix(p, option+"=") {
			return strings.TrimPrefix(p, option+"=")
		}
	}
	return ""
}

func isOption(s string) bool {
	options := []string{"primary_key", "unique", "null", "blank", "default", "index", "auto_now_add", "auto_now", "max_length", "write_only", "relation", "m2m"}
	for _, opt := range options {
		if s == opt || strings.HasPrefix(s, opt+"=") {
			return true
		}
	}
	return false
}
