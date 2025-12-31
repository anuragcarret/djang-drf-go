package fields

import (
	"testing"
)

func TestFieldOptions(t *testing.T) {
	t.Run("parses drf tag correctly", func(t *testing.T) {
		tags := `max_length=100;null;default=pending;choices=draft,pending,published`
		opts := ParseTag(tags)

		if opts.MaxLength != 100 {
			t.Errorf("expected MaxLength 100, got %d", opts.MaxLength)
		}
		if !opts.Null {
			t.Error("expected Null to be true")
		}
		if opts.Default != "pending" {
			t.Errorf("expected Default 'pending', got %v", opts.Default)
		}
		expectedChoices := []string{"draft", "pending", "published"}
		if len(opts.Choices) != len(expectedChoices) {
			t.Fatalf("expected %d choices, got %d", len(expectedChoices), len(opts.Choices))
		}
		for i, v := range expectedChoices {
			if opts.Choices[i] != v {
				t.Errorf("at index %d: expected %s, got %s", i, v, opts.Choices[i])
			}
		}
	})

	t.Run("handles boolean flags", func(t *testing.T) {
		tags := `primary_key;unique;index;blank;auto_now;auto_now_add`
		opts := ParseTag(tags)

		if !opts.PrimaryKey {
			t.Error("PrimaryKey false")
		}
		if !opts.Unique {
			t.Error("Unique false")
		}
		if !opts.Index {
			t.Error("Index false")
		}
		if !opts.Blank {
			t.Error("Blank false")
		}
		if !opts.AutoNow {
			t.Error("AutoNow false")
		}
		if !opts.AutoNowAdd {
			t.Error("AutoNowAdd false")
		}
	})
}

func TestStringField(t *testing.T) {
	t.Run("validates max_length", func(t *testing.T) {
		f := &StringField{
			BaseField: BaseField{options: &FieldOptions{MaxLength: 5}},
		}

		if err := f.Validate("hello"); err != nil {
			t.Errorf("expected no error for 'hello', got %v", err)
		}

		if err := f.Validate("hello world"); err == nil {
			t.Error("expected error for exceeding max_length")
		}
	})

	t.Run("validates blank", func(t *testing.T) {
		f := &StringField{
			BaseField: BaseField{options: &FieldOptions{Blank: false}},
		}
		if err := f.Validate(""); err == nil {
			t.Error("expected error for empty string when Blank=false")
		}

		f2 := &StringField{
			BaseField: BaseField{options: &FieldOptions{Blank: true}},
		}
		if err := f2.Validate(""); err != nil {
			t.Error("expected no error for empty string when Blank=true")
		}
	})
}

func TestIntField(t *testing.T) {
	t.Run("validates numeric types", func(t *testing.T) {
		f := &IntField{}
		if err := f.Validate(42); err != nil {
			t.Errorf("expected no error for int, got %v", err)
		}
		if err := f.Validate("42"); err == nil {
			t.Error("expected error for string in IntField")
		}
	})
}
