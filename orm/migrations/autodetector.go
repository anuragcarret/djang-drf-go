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
	for tableName, model := range models {
		if !dbTableSet[tableName] {
			// Table missing - CreateTable
			ops = append(ops, a.createTableOp(tableName, model))
		} else {
			// Table exists - Check for new columns
			newOps, err := a.detectColumnChanges(tableName, model)
			if err != nil {
				return nil, err
			}
			ops = append(ops, newOps...)
		}

		// Detect Many-to-Many through tables
		m2mOps := a.detectM2MChanges(tableName, model, dbTableSet)
		ops = append(ops, m2mOps...)
	}

	return ops, nil
}

func (a *Autodetector) detectM2MChanges(tableName string, model interface{}, dbTableSet map[string]bool) []Operation {
	var ops []Operation
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("drf")
		m2mTable := getOptionValue(tag, "m2m")
		if m2mTable == "" {
			continue
		}

		if !dbTableSet[m2mTable] {
			toCol := getOptionValue(tag, "to")
			fromCol := getOptionValue(tag, "from")
			if toCol == "" {
				toCol = "to_id"
			}
			if fromCol == "" {
				fromCol = "from_id"
			}

			// Find what table we are relating to.
			// For simplicity in this version, we assume the field's slice element type is the model.
			// In real Django, it's 'to=othermodel'

			fields := map[string]string{
				"id":    "SERIAL PRIMARY KEY",
				toCol:   "INTEGER NOT NULL",
				fromCol: "INTEGER NOT NULL",
			}

			// We can try to guess the related table from the tag 'to' or the field type
			// For now, let's keep it simple as requested
			ops = append(ops, &CreateTable{Name: m2mTable, Fields: fields})
			dbTableSet[m2mTable] = true // Avoid duplicate creation if multiple models link to same table
		}
	}
	return ops
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

		parts := strings.Split(tag, ";")
		colName := parts[0]

		if strings.Contains(colName, "=") || isOption(colName) {
			colName = toSnakeCase(f.Name)
		}

		// Type mapping
		dbType := "TEXT"
		if f.Type.Kind() == reflect.Int || f.Type.Kind() == reflect.Uint64 {
			dbType = "INTEGER"
		} else if f.Type.String() == "time.Time" || f.Type.String() == "*time.Time" {
			dbType = "TIMESTAMP WITH TIME ZONE"
		} else if f.Type.Kind() == reflect.Bool {
			dbType = "BOOLEAN"
		}

		// Constraints
		isPK := hasOption(tag, "primary_key")
		isUnique := hasOption(tag, "unique")
		isNull := hasOption(tag, "null")
		defaultValue := getOptionValue(tag, "default")
		fk := getOptionValue(tag, "foreign_key")
		o2o := getOptionValue(tag, "one_to_one")

		if isPK {
			dbType = "SERIAL PRIMARY KEY"
		} else {
			if fk != "" || o2o != "" {
				rel := fk
				if o2o != "" {
					rel = o2o
				}
				constraint := "NOT NULL"
				if o2o != "" {
					constraint = "UNIQUE NOT NULL"
				}

				relParts := strings.Split(rel, ".")
				if len(relParts) == 2 {
					dbType = fmt.Sprintf("INTEGER %s REFERENCES %s(%s)", constraint, relParts[0], relParts[1])
				}
			} else {
				if isUnique {
					dbType += " UNIQUE"
				}
				if !isNull {
					dbType += " NOT NULL"
				}
				if defaultValue != "" {
					dbType += " DEFAULT " + defaultValue
				}
			}
		}

		fields[colName] = dbType
	}
}

func isOption(s string) bool {
	options := []string{"null", "unique", "primary_key", "index", "auto_increment", "blank", "default", "foreign_key", "m2m"}
	for _, opt := range options {
		if s == opt {
			return true
		}
	}
	return false
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

func getOptionValue(tag, option string) string {
	parts := strings.Split(tag, ";")
	for _, p := range parts {
		if strings.HasPrefix(p, option+"=") {
			return strings.TrimPrefix(p, option+"=")
		}
	}
	return ""
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
		if dbType, ok := schema.Columns[colName]; !ok {
			// Column missing - AddField
			ops = append(ops, &AddField{
				TableName: tableName,
				FieldName: colName,
				FieldType: colType,
			})
		} else if dbType != colType {
			// Type or constraints changed - AlterField
			ops = append(ops, &AlterField{
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
