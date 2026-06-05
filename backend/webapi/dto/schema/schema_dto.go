package schemadto

import schemamodel "github.com/sunibgram/domain/model/schema"

type ColumnResponse struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
	IsPK     bool   `json:"is_pk"`
	IsUnique bool   `json:"is_unique"`
	IsFK     bool   `json:"is_fk"`
	FKTable  string `json:"fk_table,omitempty"`
	FKColumn string `json:"fk_column,omitempty"`
}

type ReferenceResponse struct {
	FromTable  string `json:"from_table"`
	FromColumn string `json:"from_column"`
	ToColumn   string `json:"to_column"`
}

type TableResponse struct {
	Name         string              `json:"name"`
	Columns      []ColumnResponse    `json:"columns"`
	ReferencedBy []ReferenceResponse `json:"referenced_by"`
}

type SchemaResponse struct {
	Tables []TableResponse `json:"tables"`
}

func ToSchemaResponse(tables []schemamodel.Table) SchemaResponse {
	out := SchemaResponse{Tables: make([]TableResponse, 0, len(tables))}
	for _, t := range tables {
		tr := TableResponse{
			Name:         t.Name,
			Columns:      make([]ColumnResponse, 0, len(t.Columns)),
			ReferencedBy: make([]ReferenceResponse, 0, len(t.ReferencedBy)),
		}
		for _, c := range t.Columns {
			tr.Columns = append(tr.Columns, ColumnResponse{
				Name:     c.Name,
				Type:     c.Type,
				Nullable: c.Nullable,
				IsPK:     c.IsPK,
				IsUnique: c.IsUnique,
				IsFK:     c.IsFK,
				FKTable:  c.FKTable,
				FKColumn: c.FKColumn,
			})
		}
		for _, ref := range t.ReferencedBy {
			tr.ReferencedBy = append(tr.ReferencedBy, ReferenceResponse{
				FromTable:  ref.FromTable,
				FromColumn: ref.FromColumn,
				ToColumn:   ref.ToColumn,
			})
		}
		out.Tables = append(out.Tables, tr)
	}
	return out
}
