package serializers

import (
	"fmt"
	"strings"
	"testing"
)

type TestUserSerializer struct {
	ID    uint64 `json:"id" read_only:"true"`
	Name  string `json:"name" required:"true" min_length:"3"`
	Email string `json:"email" validators:"email"`
}

func (s *TestUserSerializer) ValidateName(value string) (string, error) {
	if value == "admin" {
		return "", fmt.Errorf("name cannot be admin")
	}
	return strings.ToLower(value), nil
}

func TestBasicSerialization(t *testing.T) {
	t.Run("serializes struct to map", func(t *testing.T) {
		user := TestUserSerializer{ID: 1, Name: "John", Email: "john@example.com"}
		s := NewSerializer(&user)

		data := s.Data()
		if data["id"] != uint64(1) {
			t.Errorf("expected id 1, got %v", data["id"])
		}
		if data["name"] != "John" {
			t.Errorf("expected name 'John', got %v", data["name"])
		}
	})

	t.Run("validates input data", func(t *testing.T) {
		input := map[string]interface{}{
			"name":  "Jo",
			"email": "invalid-email",
		}
		s := NewSerializer(&TestUserSerializer{})
		err := s.Validate(input)

		if err == nil {
			t.Error("expected validation error, got nil")
		}

		errors := s.Errors()
		if _, ok := errors["name"]; !ok {
			t.Error("missing error for name")
		}
		if _, ok := errors["email"]; !ok {
			t.Error("missing error for email")
		}
	})

	t.Run("calls custom field validator", func(t *testing.T) {
		input := map[string]interface{}{
			"name":  "admin",
			"email": "test@example.com",
		}
		s := NewSerializer(&TestUserSerializer{})
		_ = s.Validate(input)

		errors := s.Errors()
		if _, ok := errors["name"]; !ok {
			t.Error("missing error for name (custom validator)")
		}
	})
}
