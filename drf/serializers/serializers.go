package serializers

import (
	"fmt"
	"net/mail"
	"reflect"
	"strconv"
	"strings"
)

// Serializer transforms data for API representation
type Serializer interface {
	Validate(data map[string]interface{}) error
	ValidatedData() map[string]interface{}
	Data() map[string]interface{}
	Errors() map[string][]string
}

// BaseSerializer provides core implementation for serializers
type BaseSerializer struct {
	instance      interface{}
	validatedData map[string]interface{}
	errors        map[string][]string
}

func NewSerializer(instance interface{}) *BaseSerializer {
	return &BaseSerializer{
		instance:      instance,
		validatedData: make(map[string]interface{}),
		errors:        make(map[string][]string),
	}
}

func (s *BaseSerializer) SetData(data map[string]interface{}) {
	s.validatedData = data
}

func (s *BaseSerializer) Data() map[string]interface{} {
	val := reflect.ValueOf(s.instance)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	data := make(map[string]interface{})
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Handle read_only/write_only is not needed for Data() (output)
		// but typically we exclude write_only fields from output
		if field.Tag.Get("write_only") == "true" {
			continue
		}

		data[jsonTag] = val.Field(i).Interface()
	}

	return data
}

func (s *BaseSerializer) Errors() map[string][]string {
	return s.errors
}

func (s *BaseSerializer) ValidatedData() map[string]interface{} {
	return s.validatedData
}

func (s *BaseSerializer) Validate(data map[string]interface{}) error {
	s.errors = make(map[string][]string)
	s.validatedData = make(map[string]interface{})

	val := reflect.ValueOf(s.instance)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		jsonName := field.Tag.Get("json")
		if jsonName == "" || jsonName == "-" {
			continue
		}

		if field.Tag.Get("read_only") == "true" {
			continue
		}

		value, ok := data[jsonName]
		if !ok {
			if field.Tag.Get("required") == "true" {
				s.addError(jsonName, "This field is required.")
			}
			continue
		}

		// Field-level validation based on tags
		if err := s.validateFieldTags(jsonName, value, field.Tag); err != nil {
			s.addError(jsonName, err.Error())
			continue
		}

		// Custom field validator: Validate<FieldName>
		validatorName := "Validate" + field.Name
		method := reflect.ValueOf(s.instance).MethodByName(validatorName)
		if method.IsValid() {
			out := method.Call([]reflect.Value{reflect.ValueOf(value)})
			if !out[1].IsNil() {
				s.addError(jsonName, out[1].Interface().(error).Error())
				continue
			}
			value = out[0].Interface()
		}

		s.validatedData[jsonName] = value
	}

	if len(s.errors) > 0 {
		return fmt.Errorf("validation failed")
	}

	return nil
}

func (s *BaseSerializer) validateFieldTags(name string, value interface{}, tag reflect.StructTag) error {
	strVal, isString := value.(string)

	// min_length
	if minLenTag := tag.Get("min_length"); minLenTag != "" && isString {
		minLen, _ := strconv.Atoi(minLenTag)
		if len(strVal) < minLen {
			return fmt.Errorf("ensure this field has at least %d characters", minLen)
		}
	}

	// validators=email
	if validators := tag.Get("validators"); validators != "" {
		for _, v := range strings.Split(validators, ",") {
			if v == "email" && isString {
				if _, err := mail.ParseAddress(strVal); err != nil {
					return fmt.Errorf("enter a valid email address")
				}
			}
		}
	}

	return nil
}

func (s *BaseSerializer) addError(field, msg string) {
	s.errors[field] = append(s.errors[field], msg)
}
