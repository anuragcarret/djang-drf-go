package migrations

import (
	"reflect"
	"testing"
	"time"

	"github.com/anuragcarret/djang-drf-go/orm/db"
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
				"username": "VARCHAR(150) UNIQUE NOT NULL",
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

func TestFieldSemanticEquality(t *testing.T) {
	detector := &Autodetector{}

	tests := []struct {
		name      string
		modelType string
		dbType    string
		expected  bool
	}{
		{
			name:      "Exact match",
			modelType: "TEXT UNIQUE NOT NULL",
			dbType:    "TEXT UNIQUE NOT NULL",
			expected:  true,
		},
		{
			name:      "Primary Key vs Introspection",
			modelType: "SERIAL PRIMARY KEY",
			dbType:    "INTEGER UNIQUE NOT NULL",
			expected:  true,
		},
		{
			name:      "Nullable mismatch",
			modelType: "TEXT UNIQUE",
			dbType:    "TEXT UNIQUE NOT NULL",
			expected:  false,
		},
		{
			name:      "Default value match",
			modelType: "BOOLEAN NOT NULL DEFAULT true",
			dbType:    "BOOLEAN NOT NULL DEFAULT true",
			expected:  true,
		},
		{
			name:      "Default value mismatch",
			modelType: "INTEGER NOT NULL DEFAULT 18",
			dbType:    "INTEGER NOT NULL DEFAULT 20",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.isFieldEqual(tt.modelType, tt.dbType)
			if got != tt.expected {
				t.Errorf("%s: isFieldEqual(%q, %q) = %v, expected %v", tt.name, tt.modelType, tt.dbType, got, tt.expected)
			}
		})
	}
}

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

func TestComprehensiveTypes(t *testing.T) {
	detector := &Autodetector{}

	type AllTypesModel struct {
		BoolField     bool                   `drf:"bool_field"`
		SmallIntField int16                  `drf:"smallint_field"`
		IntField      int32                  `drf:"int_field"`
		BigIntField   int64                  `drf:"bigint_field"`
		NumericField  float64                `drf:"numeric_field;type=numeric"`
		DoubleField   float64                `drf:"double_field"`
		VarcharField  string                 `drf:"varchar_field;max_length=255"`
		TextField     string                 `drf:"text_field"`
		DateField     time.Time              `drf:"date_field;type=date"`
		TimeField     time.Time              `drf:"time_field;type=time"`
		IntervalField string                 `drf:"interval_field;type=interval"`
		UUIDField     string                 `drf:"uuid_field;type=uuid"`
		ByteaField    []byte                 `drf:"bytea_field"`
		JSONBField    map[string]interface{} `drf:"jsonb_field"`
		ArrayField    []int32                `drf:"array_field"`
		TsVectorField string                 `drf:"ts_vector_field;type=tsvector"`
	}

	expected := map[string]string{
		"bool_field":      "BOOLEAN NOT NULL",
		"smallint_field":  "SMALLINT NOT NULL",
		"int_field":       "INTEGER NOT NULL",
		"bigint_field":    "BIGINT NOT NULL",
		"numeric_field":   "NUMERIC NOT NULL",
		"double_field":    "DOUBLE PRECISION NOT NULL",
		"varchar_field":   "VARCHAR(255) NOT NULL",
		"text_field":      "TEXT NOT NULL",
		"date_field":      "DATE NOT NULL",
		"time_field":      "TIME NOT NULL",
		"interval_field":  "INTERVAL NOT NULL",
		"uuid_field":      "UUID NOT NULL",
		"bytea_field":     "BYTEA NOT NULL",
		"jsonb_field":     "JSONB NOT NULL",
		"array_field":     "INTEGER[] NOT NULL",
		"ts_vector_field": "TSVECTOR NOT NULL",
	}

	fields := make(map[string]string)
	detector.collectFields(reflect.TypeOf(AllTypesModel{}), fields)

	for name, want := range expected {
		if got, ok := fields[name]; !ok {
			t.Errorf("Field %s missing", name)
		} else if got != want {
			t.Errorf("Field %s: got %q, want %q", name, got, want)
		}
	}
}

func TestDetectRemovedField(t *testing.T) {
	detector := &Autodetector{}

	type ModelWithLessFields struct {
		ID uint64 `drf:"id;primary_key"`
	}

	schema := &db.TableInfo{
		Name: "test_table",
		Columns: map[string]string{
			"id":        "INTEGER UNIQUE NOT NULL",
			"old_field": "TEXT",
		},
	}

	// detectColumnChangesUses uses collector.collectFields
	// We check if "old_field" is detected for removal
	ops, err := detector.detectColumnChangesInternal("test_table", &ModelWithLessFields{}, schema)
	if err != nil {
		t.Fatalf("detectColumnChangesInternal failed: %v", err)
	}

	found := false
	for _, op := range ops {
		if rop, ok := op.(*RemoveField); ok {
			if rop.FieldName == "old_field" {
				found = true
				break
			}
		}
	}

	if !found {
		t.Errorf("Expected RemoveField operation for 'old_field', but not found in %v", ops)
	}
}
