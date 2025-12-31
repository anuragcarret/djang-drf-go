package orm

import "time"

// Model is the base interface for all models.
type Model interface {
	GetTableName() string
}

// BaseModel provides common fields similar to Django.
type BaseModel struct {
	ID        uint      `orm:"primary_key;auto_increment" json:"id"`
	CreatedAt time.Time `orm:"auto_now_add" json:"created_at"`
	UpdatedAt time.Time `orm:"auto_now" json:"updated_at"`
}

// GetTableName returns the table name based on conventions (to be overridden).
func (b *BaseModel) GetTableName() string {
	return "" // Framework will use reflection to determine if empty
}
