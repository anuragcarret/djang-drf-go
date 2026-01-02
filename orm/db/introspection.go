package db

import (
	"fmt"
	"strings"
)

// TableInfo represents metadata about a table
type TableInfo struct {
	Name    string
	Columns map[string]string
}

// GetTableSchema returns the current schema of a table from the database
func (db *DB) GetTableSchema(tableName string) (*TableInfo, error) {
	// Query to get columns, types, nullability, unique constraints, and max length
	query := `
		SELECT 
			c.column_name, 
			c.data_type, 
			c.is_nullable,
			c.column_default,
			c.character_maximum_length,
			c.udt_name,
			EXISTS (
				SELECT 1 
				FROM information_schema.table_constraints tc 
				JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name 
				WHERE tc.table_name = c.table_name 
				  AND kcu.column_name = c.column_name 
				  AND (tc.constraint_type = 'UNIQUE' OR tc.constraint_type = 'PRIMARY KEY')
			) as is_unique
		FROM information_schema.columns c
		WHERE c.table_name = $1
	`
	rows, err := db.conn.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols := make(map[string]string)
	for rows.Next() {
		var name, dtype, nullable string
		var isUnique bool
		var colDefault, maxLen, udtName *string
		if err := rows.Scan(&name, &dtype, &nullable, &colDefault, &maxLen, &udtName, &isUnique); err != nil {
			return nil, err
		}

		// Normalize type to match our collectFields output for comparison
		normType := strings.ToUpper(dtype)

		// Postgres specific normalizations
		switch dtype {
		case "integer":
			normType = "INTEGER"
		case "bigint":
			normType = "BIGINT"
		case "smallint":
			normType = "SMALLINT"
		case "boolean":
			normType = "BOOLEAN"
		case "text":
			normType = "TEXT"
		case "character varying":
			if maxLen != nil {
				normType = fmt.Sprintf("VARCHAR(%s)", *maxLen)
			} else {
				normType = "VARCHAR"
			}
		case "double precision":
			normType = "DOUBLE PRECISION"
		case "timestamp with time zone":
			normType = "TIMESTAMP WITH TIME ZONE"
		case "ARRAY":
			if udtName != nil {
				switch *udtName {
				case "_int4":
					normType = "INTEGER[]"
				case "_int8":
					normType = "BIGINT[]"
				case "_text":
					normType = "TEXT[]"
				case "_float8":
					normType = "DOUBLE PRECISION[]"
				case "_bool":
					normType = "BOOLEAN[]"
				default:
					normType = strings.TrimPrefix(*udtName, "_") + "[]"
				}
			}
		default:
			// Fallback to udtName for types non-standard in information_schema.columns
			if udtName != nil {
				u := strings.ToUpper(*udtName)
				if u == "JSONB" || u == "UUID" || u == "INET" || u == "TSVECTOR" || u == "INTERVAL" || u == "DATE" || u == "TIME" || u == "NUMERIC" || u == "BYTEA" || u == "HSTORE" {
					normType = u
				}
			}
		}

		if isUnique {
			normType += " UNIQUE"
		}
		if nullable == "NO" {
			normType += " NOT NULL"
		}
		if colDefault != nil {
			// Clean up default: e.g. "'active'::text" -> "active", or "true" -> "true"
			d := *colDefault
			d = strings.Split(d, "::")[0]
			d = strings.Trim(d, "'")

			// Only append if it's not a sequence (like nextval)
			if !strings.Contains(d, "nextval") {
				normType += " DEFAULT " + d
			}
		}

		cols[name] = normType
	}

	if len(cols) == 0 {
		return nil, nil // Table doesn't exist
	}

	return &TableInfo{Name: tableName, Columns: cols}, nil
}

// GetTables returns a list of all user tables in the database
func (db *DB) GetTables() ([]string, error) {
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
	`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, nil
}
