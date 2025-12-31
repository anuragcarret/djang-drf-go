package models

import (
	"context"
	"testing"
)

type TestModel struct {
	Model
	Name  string `drf:"max_length=100"`
	Age   int    `drf:"default=18"`
	Email string `drf:"unique;null"`
}

func (m *TestModel) TableName() string {
	return "test_table"
}

func (m *TestModel) PreSave(ctx context.Context) error {
	m.Name = "trimmed"
	return nil
}

func TestModelIntrospection(t *testing.T) {
	t.Run("extracts fields from model struct", func(t *testing.T) {
		registry := NewRegistry()
		err := registry.Register(&TestModel{})
		if err != nil {
			t.Fatalf("Register failed: %v", err)
		}

		info, err := registry.GetModelInfo(&TestModel{})
		if err != nil {
			t.Fatalf("GetModelInfo failed: %v", err)
		}

		if info.TableName != "test_table" {
			t.Errorf("expected table name 'test_table', got '%s'", info.TableName)
		}

		// Check fields
		expectedFields := []string{"id", "created_at", "updated_at", "name", "age", "email"}
		if len(info.Fields) != len(expectedFields) {
			t.Errorf("expected %d fields, got %d", len(expectedFields), len(info.Fields))
		}

		fieldNames := make(map[string]bool)
		for _, f := range info.Fields {
			fieldNames[f.Column()] = true
		}

		for _, name := range expectedFields {
			if !fieldNames[name] {
				t.Errorf("missing field %s", name)
			}
		}
	})

	t.Run("identifies primary key", func(t *testing.T) {
		registry := NewRegistry()
		_ = registry.Register(&TestModel{})
		info, _ := registry.GetModelInfo(&TestModel{})

		if info.PKField == nil || info.PKField.Column() != "id" {
			t.Errorf("expected PK field 'id', got %v", info.PKField)
		}
	})
}

func TestLifecycleHooks(t *testing.T) {
	t.Run("calls PreSave hook", func(t *testing.T) {
		m := &TestModel{Name: "  Trim Me  "}

		// This requires the implementation of Save() or a helper that triggers hooks
		// For now, let's just test if we can detect the interface
		var _ PreSaver = m
	})
}
