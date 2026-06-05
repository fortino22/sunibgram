// Package schemamodel holds value structs describing the live database schema
// as introspected from Postgres information_schema. These are read-only views,
// not GORM-mapped entities — there is no table behind them.
package schemamodel

// Column is a single column of a table.
type Column struct {
	Name     string
	Type     string
	Nullable bool
	IsPK     bool
	IsUnique bool
	IsFK     bool
	FKTable  string // populated when IsFK
	FKColumn string // populated when IsFK
}

// Reference is an incoming foreign key (another table pointing at this one).
type Reference struct {
	FromTable  string
	FromColumn string
	ToColumn   string
}

// Table is a single table with its columns and the FKs that point to it.
type Table struct {
	Name         string
	Columns      []Column
	ReferencedBy []Reference
}
