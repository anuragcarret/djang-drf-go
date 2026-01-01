package queryset

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/anuragcarret/djang-drf-go/core/apps"
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
	db              *db.DB
	filters         []Q
	excludes        []Q
	ordering        []string
	limit           int
	offset          int
	distinct        bool
	forUpdate       bool
	selectRelated   []string
	prefetchRelated []string
}

// NewQuerySet creates a new queryset
func NewQuerySet[T ModelInterface](database *db.DB) *QuerySet[T] {
	return &QuerySet[T]{
		db:              database,
		filters:         make([]Q, 0),
		excludes:        make([]Q, 0),
		ordering:        make([]string, 0),
		selectRelated:   make([]string, 0),
		prefetchRelated: make([]string, 0),
	}
}

// SelectRelated specifies foreign-key relationships to follow in the query
func (q *QuerySet[T]) SelectRelated(fields ...string) *QuerySet[T] {
	newQs := q.clone()
	newQs.selectRelated = append(newQs.selectRelated, fields...)
	return newQs
}

// PrefetchRelated specifies Many-to-Many or reverse relationships to fetch in separate queries
func (q *QuerySet[T]) PrefetchRelated(fields ...string) *QuerySet[T] {
	newQs := q.clone()
	newQs.prefetchRelated = append(newQs.prefetchRelated, fields...)
	return newQs
}

// clone creates a shallow copy of the queryset for immutability
func (q *QuerySet[T]) clone() *QuerySet[T] {
	newQs := *q
	// Copy slices to avoid shared state
	newQs.filters = append([]Q(nil), q.filters...)
	newQs.excludes = append([]Q(nil), q.excludes...)
	newQs.ordering = append([]string(nil), q.ordering...)
	newQs.selectRelated = append([]string(nil), q.selectRelated...)
	newQs.prefetchRelated = append([]string(nil), q.prefetchRelated...)
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
	tableName := q.getTableName()

	// Use table prefix to avoid ambiguity during joins
	query := fmt.Sprintf("SELECT %s.* FROM %s", tableName, tableName)
	args := []interface{}{}

	// Handle SELECT RELATED (JOINs)
	if len(q.selectRelated) > 0 {
		var zero T
		t := reflect.TypeOf(zero)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		for _, relField := range q.selectRelated {
			// Find field by name (case-insensitive)
			var field reflect.StructField
			found := false
			for i := 0; i < t.NumField(); i++ {
				f := t.Field(i)
				if strings.EqualFold(f.Name, relField) {
					field = f
					found = true
					break
				}
			}

			if found {
				tag := field.Tag.Get("drf")
				parts := strings.Split(tag, ";")
				localCol := parts[0]
				var fk string
				for _, p := range parts {
					if strings.HasPrefix(p, "foreign_key=") {
						fk = strings.TrimPrefix(p, "foreign_key=")
						break
					}
				}

				if fk != "" {
					fkParts := strings.Split(fk, ".")
					if len(fkParts) == 2 {
						relatedTable := fkParts[0]
						relatedCol := fkParts[1]
						query += fmt.Sprintf(" INNER JOIN %s ON %s.%s = %s.%s",
							relatedTable, tableName, localCol, relatedTable, relatedCol)
					}
				}
			}
		}
	}

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

				column, operator := q.parseLookup(k)

				switch operator {
				case "in":
					vals := reflect.ValueOf(v)
					placeholders := []string{}
					for idx := 0; idx < vals.Len(); idx++ {
						args = append(args, vals.Index(idx).Interface())
						placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)))
					}
					query += fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ", "))
				case "contains":
					query += fmt.Sprintf("%s LIKE $%d", column, len(args)+1)
					args = append(args, fmt.Sprintf("%%%v%%", v))
				case "icontains":
					query += fmt.Sprintf("%s ILIKE $%d", column, len(args)+1)
					args = append(args, fmt.Sprintf("%%%v%%", v))
				case "iexact":
					query += fmt.Sprintf("%s ILIKE $%d", column, len(args)+1)
					args = append(args, v)
				case "gt":
					query += fmt.Sprintf("%s > $%d", column, len(args)+1)
					args = append(args, v)
				case "gte":
					query += fmt.Sprintf("%s >= $%d", column, len(args)+1)
					args = append(args, v)
				case "lt":
					query += fmt.Sprintf("%s < $%d", column, len(args)+1)
					args = append(args, v)
				case "lte":
					query += fmt.Sprintf("%s <= $%d", column, len(args)+1)
					args = append(args, v)
				default:
					query += fmt.Sprintf("%s = $%d", column, len(args)+1)
					args = append(args, v)
				}
				j++
			}
		}
	}
	// ... ordering/limit/offset omitted for brevity if using replace_file_content properly ...
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

func (q *QuerySet[T]) parseLookup(key string) (string, string) {
	parts := strings.Split(key, "__")
	if len(parts) > 1 {
		return parts[0], parts[1]
	}
	return key, "exact"
}

func (q *QuerySet[T]) getTableName() string {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	instance := reflect.New(t).Interface()
	if m, ok := instance.(ModelInterface); ok {
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
			fieldFound := false
			for j := 0; j < elem.NumField(); j++ {
				f := elem.Type().Field(j)
				tag := f.Tag.Get("drf")
				parts := strings.Split(tag, ";")
				colName := parts[0]
				if colName == col || f.Name == col {
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

	// Handle PrefetchRelated after main results are fetched
	if len(q.prefetchRelated) > 0 {
		for _, fieldName := range q.prefetchRelated {
			if err := q.handlePrefetch(results, fieldName); err != nil {
				return nil, err
			}
		}
	}

	return results, nil
}

func (q *QuerySet[T]) handlePrefetch(results []T, fieldName string) error {
	if len(results) == 0 {
		return nil
	}

	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	field, found := t.FieldByName(fieldName)
	if !found {
		return fmt.Errorf("field %s not found on model %s", fieldName, t.Name())
	}

	tag := field.Tag.Get("drf")
	rel := getOptionValue(tag, "relation")
	m2m := getOptionValue(tag, "m2m")

	if rel == "" && m2m == "" {
		return nil
	}

	if m2m != "" {
		return q.handleM2M(results, fieldName, m2m, tag)
	}

	parts := strings.Split(rel, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid relation tag format: %s", rel)
	}
	return q.handleReverseFK(results, fieldName, parts[0], parts[1])
}

func (q *QuerySet[T]) handleM2M(results []T, fieldName, throughTable, tag string) error {
	toCol := getOptionValue(tag, "to")
	fromCol := getOptionValue(tag, "from")

	ids := make([]interface{}, 0)
	idMap := make(map[interface{}][]int)
	for i, res := range results {
		val := reflect.ValueOf(res)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		idField := val.FieldByName("ID")
		if !idField.IsValid() {
			continue
		}
		idVal := idField.Interface()
		ids = append(ids, idVal)
		idMap[idVal] = append(idMap[idVal], i)
	}

	if len(ids) == 0 {
		return nil
	}

	// Find related model info to get its table name and fields
	var zero T
	parentType := reflect.TypeOf(zero)
	if parentType.Kind() == reflect.Ptr {
		parentType = parentType.Elem()
	}
	field, _ := parentType.FieldByName(fieldName)
	sliceType := field.Type
	elemType := sliceType.Elem()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	relatedInstance := reflect.New(elemType).Interface()
	relatedModel, ok := relatedInstance.(ModelInterface)
	if !ok {
		return fmt.Errorf("related model does not implement ModelInterface")
	}
	relatedTable := relatedModel.TableName()

	// SELECT r.*, t.fromCol as _m2m_map_id FROM related r JOIN through t ON r.id = t.toCol WHERE t.fromCol IN (...)
	query := fmt.Sprintf("SELECT r.*, t.%s AS _m2m_map_id FROM %s r JOIN %s t ON r.id = t.%s WHERE t.%s IN (",
		fromCol, relatedTable, throughTable, toCol, fromCol)

	placeholders := make([]string, len(ids))
	for i := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	query += strings.Join(placeholders, ", ") + ")"

	rows, err := q.db.Query(query, ids...)
	if err != nil {
		return err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	for rows.Next() {
		inst := reflect.New(elemType).Elem()
		dest := make([]interface{}, len(cols))
		var m2mMapID interface{}

		for i, col := range cols {
			if col == "_m2m_map_id" {
				dest[i] = &m2mMapID
				continue
			}
			fieldFound := false
			for j := 0; j < inst.NumField(); j++ {
				f := inst.Type().Field(j)
				tag := f.Tag.Get("drf")
				parts := strings.Split(tag, ";")
				if parts[0] == col || f.Name == col {
					dest[i] = inst.Field(j).Addr().Interface()
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
			return err
		}

		if indices, ok := idMap[m2mMapID]; ok {
			for _, idx := range indices {
				parent := reflect.ValueOf(&results[idx]).Elem()
				if parent.Kind() == reflect.Ptr {
					parent = parent.Elem()
				}
				sliceField := parent.FieldByName(fieldName)
				if sliceField.IsValid() && sliceField.Kind() == reflect.Slice {
					// Handle pointer vs concrete slice
					itemToAppend := inst
					if sliceType.Elem().Kind() == reflect.Ptr {
						itemToAppend = inst.Addr()
					}
					sliceField.Set(reflect.Append(sliceField, itemToAppend))
				}
			}
		}
	}

	return nil
}

// Create inserts a new record into the database
func (q *QuerySet[T]) Create(obj T) error {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	t := val.Type()
	var tableName string
	if m, ok := interface{}(obj).(ModelInterface); ok {
		tableName = m.TableName()
	}

	fields := []string{}
	placeholders := []string{}
	values := []interface{}{}
	idx := 1

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("drf")
		if tag == "" || hasOption(tag, "auto_increment") || hasOption(tag, "relation") || hasOption(tag, "m2m") {
			continue
		}

		colName := strings.Split(tag, ";")[0]
		fields = append(fields, colName)
		placeholders = append(placeholders, fmt.Sprintf("$%d", idx))
		values = append(values, val.Field(i).Interface())
		idx++
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id",
		tableName, strings.Join(fields, ", "), strings.Join(placeholders, ", "))

	var id uint64
	err := q.db.QueryRow(query, values...).Scan(&id)
	if err != nil {
		return err
	}

	// Update the ID field in the object
	idField := val.FieldByName("ID")
	if idField.IsValid() && idField.CanSet() {
		idField.SetUint(id)
	}

	return nil
}

func (q *QuerySet[T]) handleReverseFK(results []T, fieldName, relatedTable, relatedCol string) error {
	ids := make([]interface{}, 0)
	idMap := make(map[interface{}][]int)
	for i, res := range results {
		val := reflect.ValueOf(res)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		idField := val.FieldByName("ID")
		if !idField.IsValid() {
			continue
		}
		idVal := idField.Interface()
		ids = append(ids, idVal)
		idMap[idVal] = append(idMap[idVal], i)
	}

	if len(ids) == 0 {
		return nil
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s IN (", relatedTable, relatedCol)
	placeholders := make([]string, len(ids))
	for i := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	query += strings.Join(placeholders, ", ") + ")"

	rows, err := q.db.Query(query, ids...)
	if err != nil {
		return err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	relatedModel, err := apps.Apps.GetModel(relatedTable)
	if err != nil {
		return err
	}
	relatedType := reflect.TypeOf(relatedModel)
	if relatedType.Kind() == reflect.Ptr {
		relatedType = relatedType.Elem()
	}

	for rows.Next() {
		inst := reflect.New(relatedType).Elem()
		dest := make([]interface{}, len(cols))
		for i, col := range cols {
			fieldFound := false
			for j := 0; j < inst.NumField(); j++ {
				f := inst.Type().Field(j)
				tag := f.Tag.Get("drf")
				parts := strings.Split(tag, ";")
				if parts[0] == col || f.Name == col {
					dest[i] = inst.Field(j).Addr().Interface()
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
			return err
		}

		fkField := inst.FieldByName(strings.Title(relatedCol))
		if !fkField.IsValid() {
			for j := 0; j < inst.NumField(); j++ {
				f := inst.Type().Field(j)
				tag := f.Tag.Get("drf")
				parts := strings.Split(tag, ";")
				if parts[0] == relatedCol {
					fkField = inst.Field(j)
					break
				}
			}
		}

		if fkField.IsValid() {
			fkVal := fkField.Interface()
			if indices, ok := idMap[fkVal]; ok {
				for _, idx := range indices {
					parent := reflect.ValueOf(&results[idx]).Elem()
					if parent.Kind() == reflect.Ptr {
						parent = parent.Elem()
					}
					sliceField := parent.FieldByName(fieldName)
					if sliceField.IsValid() && sliceField.Kind() == reflect.Slice {
						sliceField.Set(reflect.Append(sliceField, inst))
					}
				}
			}
		}
	}
	return nil
}

func getOptionValue(tag, option string) string {
	parts := strings.Split(tag, ";")
	for _, p := range parts {
		if strings.HasPrefix(p, option+"=") {
			return strings.TrimPrefix(p, option+"=")
		}
	}
	return ""
}

func hasOption(tag, option string) bool {
	parts := strings.Split(tag, ";")
	for _, p := range parts {
		if p == option || strings.HasPrefix(p, option+"=") {
			return true
		}
	}
	return false
}

// Count returns the number of results matching the query
func (q *QuerySet[T]) Count() (int, error) {
	sql, args := q.SQL()
	// Replace SELECT * with SELECT COUNT(*)
	sql = strings.Replace(sql, "SELECT *", "SELECT COUNT(*)", 1)

	// Remove ORDER BY if present (not needed for COUNT and might be invalid in some dialects/situations)
	if idx := strings.LastIndex(sql, " ORDER BY"); idx != -1 {
		sql = sql[:idx]
	}

	var count int
	err := q.db.QueryRow(sql, args...).Scan(&count)
	return count, err
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
