package schemarepo

import (
	"sort"

	"gorm.io/gorm"

	schemamodel "github.com/sunibgram/domain/model/schema"
	schemarepoiface "github.com/sunibgram/domain/repository/schema"
)

type SchemaRepository struct {
	readDB *gorm.DB
}

func NewSchemaRepository(writeDB *gorm.DB, readDB *gorm.DB) schemarepoiface.ISchemaRepository {
	return &SchemaRepository{readDB: readDB}
}

type columnRow struct {
	TableName  string
	ColumnName string
	DataType   string
	IsNullable string
	Position   int
}

type keyRow struct {
	TableName      string
	ColumnName     string
	ConstraintType string // PRIMARY KEY | UNIQUE
}

type fkRow struct {
	TableName        string
	ColumnName       string
	ForeignTableName string
	ForeignColumn    string
}

// GetSchema reads columns, primary/unique keys and foreign keys from
// information_schema for the public schema and assembles them per table.
func (r *SchemaRepository) GetSchema() ([]schemamodel.Table, error) {
	var cols []columnRow
	if err := r.readDB.Raw(`
		SELECT table_name AS table_name,
		       column_name AS column_name,
		       data_type AS data_type,
		       is_nullable AS is_nullable,
		       ordinal_position AS position
		FROM information_schema.columns
		WHERE table_schema = 'public'
		ORDER BY table_name, ordinal_position
	`).Scan(&cols).Error; err != nil {
		return nil, err
	}

	var keys []keyRow
	if err := r.readDB.Raw(`
		SELECT tc.table_name AS table_name,
		       kcu.column_name AS column_name,
		       tc.constraint_type AS constraint_type
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
		  ON tc.constraint_name = kcu.constraint_name
		 AND tc.table_schema = kcu.table_schema
		WHERE tc.table_schema = 'public'
		  AND tc.constraint_type IN ('PRIMARY KEY', 'UNIQUE')
	`).Scan(&keys).Error; err != nil {
		return nil, err
	}

	var fks []fkRow
	if err := r.readDB.Raw(`
		SELECT tc.table_name AS table_name,
		       kcu.column_name AS column_name,
		       ccu.table_name AS foreign_table_name,
		       ccu.column_name AS foreign_column
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
		  ON tc.constraint_name = kcu.constraint_name
		 AND tc.table_schema = kcu.table_schema
		JOIN information_schema.constraint_column_usage ccu
		  ON tc.constraint_name = ccu.constraint_name
		 AND tc.table_schema = ccu.table_schema
		WHERE tc.table_schema = 'public'
		  AND tc.constraint_type = 'FOREIGN KEY'
	`).Scan(&fks).Error; err != nil {
		return nil, err
	}

	// Index keys and FKs by (table,column) for assembly.
	pk := map[string]bool{}
	uk := map[string]bool{}
	for _, k := range keys {
		ck := k.TableName + "." + k.ColumnName
		if k.ConstraintType == "PRIMARY KEY" {
			pk[ck] = true
		} else {
			uk[ck] = true
		}
	}
	fkByCol := map[string]fkRow{}
	for _, f := range fks {
		fkByCol[f.TableName+"."+f.ColumnName] = f
	}

	// Build tables preserving column order from the columns query.
	byName := map[string]*schemamodel.Table{}
	order := []string{}
	for _, c := range cols {
		t, ok := byName[c.TableName]
		if !ok {
			t = &schemamodel.Table{Name: c.TableName}
			byName[c.TableName] = t
			order = append(order, c.TableName)
		}
		ck := c.TableName + "." + c.ColumnName
		col := schemamodel.Column{
			Name:     c.ColumnName,
			Type:     c.DataType,
			Nullable: c.IsNullable == "YES",
			IsPK:     pk[ck],
			IsUnique: uk[ck],
		}
		if f, ok := fkByCol[ck]; ok {
			col.IsFK = true
			col.FKTable = f.ForeignTableName
			col.FKColumn = f.ForeignColumn
		}
		t.Columns = append(t.Columns, col)
	}

	// Invert FKs into referenced_by on the target table.
	for _, f := range fks {
		if t, ok := byName[f.ForeignTableName]; ok {
			t.ReferencedBy = append(t.ReferencedBy, schemamodel.Reference{
				FromTable:  f.TableName,
				FromColumn: f.ColumnName,
				ToColumn:   f.ForeignColumn,
			})
		}
	}

	sort.Strings(order)
	out := make([]schemamodel.Table, 0, len(order))
	for _, n := range order {
		t := byName[n]
		sort.Slice(t.ReferencedBy, func(i, j int) bool {
			if t.ReferencedBy[i].FromTable != t.ReferencedBy[j].FromTable {
				return t.ReferencedBy[i].FromTable < t.ReferencedBy[j].FromTable
			}
			return t.ReferencedBy[i].FromColumn < t.ReferencedBy[j].FromColumn
		})
		out = append(out, *t)
	}
	return out, nil
}
