package db

// TableInfo represents metadata about a table
type TableInfo struct {
	Name    string
	Columns map[string]string
}

// GetTableSchema returns the current schema of a table from the database
func (db *DB) GetTableSchema(tableName string) (*TableInfo, error) {
	query := `
		SELECT column_name, data_type, character_maximum_length, is_nullable
		FROM information_schema.columns
		WHERE table_name = $1
	`
	rows, err := db.conn.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols := make(map[string]string)
	for rows.Next() {
		var name, dtype, nullable string
		var maxLen interface{}
		if err := rows.Scan(&name, &dtype, &maxLen, &nullable); err != nil {
			return nil, err
		}

		// Simplify types for comparison
		cols[name] = dtype
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
