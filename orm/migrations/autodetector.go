package migrations

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/anuragcarret/djang-drf-go/core/apps"
	"github.com/anuragcarret/djang-drf-go/orm/db"
)

// Autodetector compares current models with DB schema
type Autodetector struct {
	db *db.DB
}

func NewAutodetector(database *db.DB) *Autodetector {
	return &Autodetector{db: database}
}

// Changes detects missing tables and columns
func (a *Autodetector) Changes() ([]Operation, error) {
	var ops []Operation

	dbTables, err := a.db.GetTables()
	if err != nil {
		return nil, err
	}
	dbTableSet := make(map[string]bool)
	for _, t := range dbTables {
		dbTableSet[t] = true
	}

	models := apps.Apps.GetAllModels()
	fmt.Printf("Autodetector: detecting changes for %d models: %v\n", len(models), reflect.ValueOf(models).MapKeys())
	for tableName, model := range models {
		if !dbTableSet[tableName] {
			// Table missing - CreateTable
			ops = append(ops, a.createTableOp(tableName, model))
		} else {
			// Table exists - Check for new columns
			// This is a simplified version - only AddField
			newOps, err := a.detectColumnChanges(tableName, model)
			if err != nil {
				return nil, err
			}
			ops = append(ops, newOps...)
		}
	}

	return ops, nil
}

func (a *Autodetector) createTableOp(name string, model interface{}) Operation {
	fields := make(map[string]string)
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	a.collectFields(v.Type(), fields)

	return &CreateTable{Name: name, Fields: fields}
}

func (a *Autodetector) collectFields(t reflect.Type, fields map[string]string) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		// Handle embedded fields (like models.Model)
		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			a.collectFields(f.Type, fields)
			continue
		}

		tag := f.Tag.Get("drf")
		if tag == "" {
			tag = f.Tag.Get("json")
		}
		if tag == "-" || tag == "" {
			continue
		}

		// Split tag to extract column name and options
		parts := strings.Split(tag, ";")
		colName := parts[0]

		// If the first part contains an '=' or is a known option, it's not a column name
		if strings.Contains(colName, "=") || colName == "null" || colName == "unique" || colName == "primary_key" || colName == "index" {
			colName = toSnakeCase(f.Name)
		}

		// Very simplified type mapping
		dbType := "TEXT"

		isPK := false
		for _, p := range parts {
			if p == "primary_key" {
				isPK = true
				break
			}
		}

		if isPK {
			dbType = "SERIAL PRIMARY KEY"
		} else if f.Type.Kind() == reflect.Int || f.Type.Kind() == reflect.Uint64 {
			dbType = "INTEGER"
		} else if f.Type.String() == "time.Time" {
			dbType = "TIMESTAMP WITH TIME ZONE"
		} else if f.Type.Kind() == reflect.Bool {
			dbType = "BOOLEAN"
		}

		fields[colName] = dbType
	}
}

func (a *Autodetector) detectColumnChanges(tableName string, model interface{}) ([]Operation, error) {
	var ops []Operation
	schema, err := a.db.GetTableSchema(tableName)
	if err != nil {
		return nil, err
	}
	if schema == nil {
		return nil, nil // Should not happen if caller checked
	}

	modelFields := make(map[string]string)
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	a.collectFields(v.Type(), modelFields)

	for colName, colType := range modelFields {
		if _, ok := schema.Columns[colName]; !ok {
			// Column missing - AddField
			ops = append(ops, &AddField{
				TableName: tableName,
				FieldName: colName,
				FieldType: colType,
			})
		}
	}

	return ops, nil
}

func toSnakeCase(s string) string {
	res := ""
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			res += "_"
		}
		res += strings.ToLower(string(r))
	}
	return res
}
