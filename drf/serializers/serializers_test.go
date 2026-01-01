package serializers

import (
	"testing"
)

type MockUser struct {
	ID       uint64 `drf:"id;primary_key"`
	Username string `drf:"username;max_length=10"`
	Password string `drf:"password;write_only"`
	Email    string `drf:"email"`
}

func TestSerialize(t *testing.T) {
	t.Run("basic serialization handles write_only", func(t *testing.T) {
		user := &MockUser{ID: 1, Username: "test", Password: "secret", Email: "test@ex.com"}
		data := Serialize(user, 0).(map[string]interface{})

		if _, ok := data["password"]; ok {
			t.Error("write_only field 'password' should not be serialized")
		}
		if data["username"] != "test" {
			t.Errorf("expected username 'test', got %v", data["username"])
		}
	})
}

type MockRel struct {
	ID   uint64    `drf:"id;primary_key"`
	Name string    `drf:"name"`
	Subs []MockRel `drf:"relation=mock_rel.parent_id"`
}

func TestNestedSerialization(t *testing.T) {
	t.Run("serializes nested relations with depth", func(t *testing.T) {
		child := MockRel{ID: 2, Name: "child"}
		parent := MockRel{ID: 1, Name: "parent", Subs: []MockRel{child}}

		// Depth 0: should NOT include Subs
		flat := Serialize(parent, 0).(map[string]interface{})
		if _, ok := flat["Subs"]; ok {
			t.Error("Subs should not be serialized at depth 0")
		}

		// Depth 1: SHOULD include Subs
		nested := Serialize(parent, 1).(map[string]interface{})
		subs, ok := nested["Subs"].([]interface{})
		if !ok || len(subs) != 1 {
			t.Fatalf("Expected 1 nested sub, got %v", nested["Subs"])
		}
		sub := subs[0].(map[string]interface{})
		if sub["name"] != "child" {
			t.Errorf("Expected nested name 'child', got %v", sub["name"])
		}
	})
}

func TestBaseSerializer_Validation(t *testing.T) {
	t.Run("validates max_length correctly", func(t *testing.T) {
		serializer := NewSerializer(&MockUser{})
		data := map[string]interface{}{
			"username": "too_long_username",
			"email":    "test@example.com",
		}

		if serializer.IsValid(data) {
			t.Error("Expected IsValid to be false for too long username")
		}

		if err, ok := serializer.Errors()["username"]; !ok || err == "" {
			t.Error("Expected error message for username field")
		}
	})

	t.Run("validates required fields", func(t *testing.T) {
		serializer := NewSerializer(&MockUser{})
		data := map[string]interface{}{
			"email": "test@example.com",
		}

		if serializer.IsValid(data) {
			t.Error("Expected IsValid to be false when required field 'username' is missing")
		}
	})
}
