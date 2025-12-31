package queryset

import (
	"fmt"
	"reflect"

	"github.com/anuragcarret/djang-drf-go/orm/db"
)

// Q represents query parameters for filtering
type Q map[string]interface{}

// ModelInterface mirrors orm.ModelInterface to avoid circular dependency
type ModelInterface interface {
	TableName() string
}

// QuerySet is a lazy, chainable query builder
type QuerySet[T ModelInterface] struct {
	db        *db.DB
	filters   []Q
	excludes  []Q
	ordering  []string
	limit     int
	offset    int
	distinct  bool
	forUpdate bool
}

// NewQuerySet creates a new queryset
func NewQuerySet[T ModelInterface](database *db.DB) *QuerySet[T] {
	return &QuerySet[T]{
		db:       database,
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

// SQL returns the generated SQL query and arguments
func (q *QuerySet[T]) SQL() (string, []interface{}) {
	var zero T
	modelType := reflect.TypeOf(zero)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	tableName := q.getTableName()

	query := fmt.Sprintf("SELECT * FROM %s", tableName)
	args := []interface{}{}

	if len(q.filters) > 0 {
		query += " WHERE "
		for i, filter := range q.filters {
			if i > 0 {
				query += " AND "
			}
			j := 0
			for k, v := range filter {
				if j > 0 {
					query += " AND "
				}
				// Simple "=" for now. Lookups like __gte will be handled later.
				query += fmt.Sprintf("%s = $%d", k, len(args)+1)
				args = append(args, v)
				j++
			}
		}
	}

	if len(q.ordering) > 0 {
		query += " ORDER BY "
		for i, field := range q.ordering {
			if i > 0 {
				query += ", "
			}
			if field[0] == '-' {
				query += fmt.Sprintf("%s DESC", field[1:])
			} else {
				query += field
			}
		}
	}

	if q.limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", q.limit)
	}
	if q.offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", q.offset)
	}

	return query, args
}

func (q *QuerySet[T]) getTableName() string {
	var zero T
	// Hack: if it's a pointer, we need an instance to call TableName
	// In a real implementation, we'd use the Registry to get metadata by Type
	val := reflect.New(reflect.TypeOf(zero).Elem()).Interface()
	if m, ok := val.(ModelInterface); ok {
		return m.TableName()
	}
	return ""
}

// All returns all matching records (Terminal operation)
func (q *QuerySet[T]) All() ([]T, error) {
	query, args := q.SQL()
	rows, err := q.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []T
	for rows.Next() {
		var item T
		// If T is a pointer type, we need to allocate it
		val := reflect.ValueOf(&item).Elem()
		if val.Kind() == reflect.Ptr {
			val.Set(reflect.New(val.Type().Elem()))
		}

		elem := val
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}

		dest := make([]interface{}, len(cols))
		for i, col := range cols {
			// Find field in struct by tag or name
			fieldFound := false
			for j := 0; j < elem.NumField(); j++ {
				f := elem.Type().Field(j)
				tag := f.Tag.Get("drf")
				if tag == "" {
					tag = f.Tag.Get("json")
				}
				if tag == col || f.Name == col {
					dest[i] = elem.Field(j).Addr().Interface()
					fieldFound = true
					break
				}
			}
			if !fieldFound {
				var dummy interface{}
				dest[i] = &dummy
			}
		}

		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, nil
}

// Get returns exactly one record (Terminal operation)
func (q *QuerySet[T]) Get(params ...Q) (T, error) {
	var zero T
	qs := q
	if len(params) > 0 {
		qs = q.Filter(params...)
	}

	results, err := qs.Limit(2).All()
	if err != nil {
		return zero, err
	}

	if len(results) == 0 {
		return zero, fmt.Errorf("does not exist")
	}
	if len(results) > 1 {
		return zero, fmt.Errorf("multiple objects returned")
	}

	return results[0], nil
}
