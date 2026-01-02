package migrations

import (
	"reflect"
	"testing"
)

type TestModel struct {
	ID       uint64 `drf:"id;primary_key"`
	Username string `drf:"username;unique;max_length=150"`
	Email    string `drf:"email;unique;null"`
	Bio      string `drf:"bio;null;blank"`
	Age      int    `drf:"age;default=18"`
}

func (t *TestModel) TableName() string { return "test_model" }

type RelatedModel struct {
	ID uint64 `drf:"id;primary_key"`
}

func (r *RelatedModel) TableName() string { return "related_model" }

type FKModel struct {
	ID      uint64        `drf:"id;primary_key"`
	Related *RelatedModel `drf:"related_id;foreign_key=related_model.id"`
}

func (f *FKModel) TableName() string { return "fk_model" }

type O2OModel struct {
	ID      uint64        `drf:"id;primary_key"`
	Related *RelatedModel `drf:"related_id;one_to_one=related_model.id"`
}

func (o *O2OModel) TableName() string { return "o2o_model" }

func TestCollectFields(t *testing.T) {
	detector := &Autodetector{}

	tests := []struct {
		name     string
		model    interface{}
		expected map[string]string
	}{
		{
			name:  "Basic fields with constraints",
			model: &TestModel{},
			expected: map[string]string{
				"id":       "SERIAL PRIMARY KEY",
				"username": "TEXT UNIQUE NOT NULL",
				"email":    "TEXT UNIQUE",
				"bio":      "TEXT",
				"age":      "INTEGER NOT NULL DEFAULT 18",
			},
		},
		{
			name:  "Foreign Key field",
			model: &FKModel{},
			expected: map[string]string{
				"id":         "SERIAL PRIMARY KEY",
				"related_id": "INTEGER NOT NULL REFERENCES related_model(id)",
			},
		},
		{
			name:  "One to One field",
			model: &O2OModel{},
			expected: map[string]string{
				"id":         "SERIAL PRIMARY KEY",
				"related_id": "INTEGER UNIQUE NOT NULL REFERENCES related_model(id)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := make(map[string]string)
			v := reflect.TypeOf(tt.model)
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}
			detector.collectFields(v, fields)

			for k, v := range tt.expected {
				if fields[k] != v {
					t.Errorf("Field %s: expected %s, got %s", k, v, fields[k])
				}
			}
			if len(fields) != len(tt.expected) {
				t.Errorf("Expected %d fields, got %d: %v", len(tt.expected), len(fields), fields)
			}
		})
	}
}

type M2MModel struct {
	ID        uint64     `drf:"id;primary_key"`
	Followers []M2MModel `drf:"m2m=test_m2m_through;to=to_id;from=from_id"`
}

func (m *M2MModel) TableName() string { return "m2m_model" }

func TestM2MThroughTableDetection(t *testing.T) {
	// Mock DB that returns no tables
	// Since we can't easily mock db.DB without an interface,
	// we'll just test that detectM2MChanges returns the expected operation.
	detector := &Autodetector{}
	dbTableSet := make(map[string]bool)

	model := &M2MModel{}
	ops := detector.detectM2MChanges("m2m_model", model, dbTableSet)

	if len(ops) != 1 {
		t.Fatalf("Expected 1 operation, got %d", len(ops))
	}

	ct, ok := ops[0].(*CreateTable)
	if !ok {
		t.Fatalf("Expected CreateTable operation, got %T", ops[0])
	}

	if ct.Name != "test_m2m_through" {
		t.Errorf("Expected through table name test_m2m_through, got %s", ct.Name)
	}

	if _, ok := ct.Fields["to_id"]; !ok {
		t.Errorf("Expected field to_id in through table")
	}
	if _, ok := ct.Fields["from_id"]; !ok {
		t.Errorf("Expected field from_id in through table")
	}
}
