package db

// TableInfo represents metadata about a table
type TableInfo struct {
	Name    string
	Columns map[string]string
}

// GetTableSchema returns the current schema of a table from the database
func (db *DB) GetTableSchema(tableName string) (*TableInfo, error) {
	// Query to get columns, types, nullability and unique constraints
	query := `
		SELECT 
			c.column_name, 
			c.data_type, 
			c.is_nullable,
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
		if err := rows.Scan(&name, &dtype, &nullable, &isUnique); err != nil {
			return nil, err
		}

		// Normalize type to match our collectFields output for comparison
		normType := "TEXT"
		if dtype == "integer" || dtype == "bigint" {
			normType = "INTEGER"
		} else if dtype == "timestamp with time zone" {
			normType = "TIMESTAMP WITH TIME ZONE"
		} else if dtype == "boolean" {
			normType = "BOOLEAN"
		}

		if isUnique { // Simplification: we treat PK as UNIQUE too for this comparison
			normType += " UNIQUE"
		}
		if nullable == "NO" {
			normType += " NOT NULL"
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
