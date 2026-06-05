package schemarepoiface

import schemamodel "github.com/sunibgram/domain/model/schema"

// ISchemaRepository introspects the live Postgres schema (read-only).
type ISchemaRepository interface {
	GetSchema() ([]schemamodel.Table, error)
}
