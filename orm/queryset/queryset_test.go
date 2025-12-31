package queryset

import (
	"testing"
)

type MockModel struct {
	ID   uint64
	Name string
}

func (m *MockModel) TableName() string { return "mock_table" }
func (m *MockModel) Meta() interface{} { return nil }

func TestQuerySetChainability(t *testing.T) {
	t.Run("chains filters and ordering", func(t *testing.T) {
		qs := NewQuerySet[*MockModel]()

		qs = qs.Filter(Q{"is_active": true}).
			Exclude(Q{"is_staff": true}).
			OrderBy("-created_at").
			Limit(10).
			Offset(20)

		if len(qs.filters) != 1 {
			t.Errorf("expected 1 filter, got %d", len(qs.filters))
		}
		if len(qs.excludes) != 1 {
			t.Errorf("expected 1 exclude, got %d", len(qs.excludes))
		}
		if len(qs.ordering) != 1 || qs.ordering[0] != "-created_at" {
			t.Errorf("expected ordering ['-created_at'], got %v", qs.ordering)
		}
		if qs.limit != 10 {
			t.Errorf("expected limit 10, got %d", qs.limit)
		}
		if qs.offset != 20 {
			t.Errorf("expected offset 20, got %d", qs.offset)
		}
	})

	t.Run("Filter appends to existing filters", func(t *testing.T) {
		qs := NewQuerySet[*MockModel]().
			Filter(Q{"a": 1}).
			Filter(Q{"b": 2})

		if len(qs.filters) != 2 {
			t.Errorf("expected 2 filters, got %d", len(qs.filters))
		}
	})
}

func TestQObjects(t *testing.T) {
	t.Run("merges multiple Q objects in Filter", func(t *testing.T) {
		qs := NewQuerySet[*MockModel]().Filter(Q{"a": 1}, Q{"b": 2})

		if len(qs.filters) != 1 {
			t.Fatalf("expected 1 filter entry, got %d", len(qs.filters))
		}

		filter := qs.filters[0]
		if filter["a"] != 1 || filter["b"] != 2 {
			t.Errorf("Q objects were not merged correctly: %v", filter)
		}
	})
}
