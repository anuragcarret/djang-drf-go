package queryset

import (
	"fmt"
)

// Q represents query parameters for filtering
type Q map[string]interface{}

// ModelInterface mirrors orm.ModelInterface to avoid circular dependency
// In actual implementation, we might need a common core or shared interface package
type ModelInterface interface {
	TableName() string
}

// QuerySet is a lazy, chainable query builder
type QuerySet[T ModelInterface] struct {
	filters   []Q
	excludes  []Q
	ordering  []string
	limit     int
	offset    int
	distinct  bool
	forUpdate bool
}

// NewQuerySet creates a new queryset
func NewQuerySet[T ModelInterface]() *QuerySet[T] {
	return &QuerySet[T]{
		filters:  make([]Q, 0),
		excludes: make([]Q, 0),
		ordering: make([]string, 0),
	}
}

// clone creates a shallow copy of the queryset for immutability
func (q *QuerySet[T]) clone() *QuerySet[T] {
	newQs := *q
	// Copy slices to avoid shared state
	newQs.filters = append([]Q(nil), q.filters...)
	newQs.excludes = append([]Q(nil), q.excludes...)
	newQs.ordering = append([]string(nil), q.ordering...)
	return &newQs
}

// Filter adds positive filter criteria
func (q *QuerySet[T]) Filter(params ...Q) *QuerySet[T] {
	newQs := q.clone()
	if len(params) == 0 {
		return newQs
	}

	merged := Q{}
	for _, p := range params {
		for k, v := range p {
			merged[k] = v
		}
	}
	newQs.filters = append(newQs.filters, merged)
	return newQs
}

// Exclude adds negative filter criteria
func (q *QuerySet[T]) Exclude(params ...Q) *QuerySet[T] {
	newQs := q.clone()
	if len(params) == 0 {
		return newQs
	}

	merged := Q{}
	for _, p := range params {
		for k, v := range p {
			merged[k] = v
		}
	}
	newQs.excludes = append(newQs.excludes, merged)
	return newQs
}

// OrderBy sets the sort order
func (q *QuerySet[T]) OrderBy(fields ...string) *QuerySet[T] {
	newQs := q.clone()
	newQs.ordering = fields
	return newQs
}

// Limit sets the maximum number of records
func (q *QuerySet[T]) Limit(n int) *QuerySet[T] {
	newQs := q.clone()
	newQs.limit = n
	return newQs
}

// Offset sets the number of records to skip
func (q *QuerySet[T]) Offset(n int) *QuerySet[T] {
	newQs := q.clone()
	newQs.offset = n
	return newQs
}

// Count returns the number of matching records (Terminal operation)
func (q *QuerySet[T]) Count() (int64, error) {
	// TODO: SQL generation and execution
	return 0, nil
}

// All returns all matching records (Terminal operation)
func (q *QuerySet[T]) All() ([]T, error) {
	// TODO: SQL generation and execution
	return nil, nil
}

// Get returns exactly one record (Terminal operation)
func (q *QuerySet[T]) Get(params ...Q) (T, error) {
	var zero T
	qs := q
	if len(params) > 0 {
		qs = q.Filter(params...)
	}
	_ = qs
	// TODO: SQL generation and execution
	return zero, fmt.Errorf("not implemented")
}
