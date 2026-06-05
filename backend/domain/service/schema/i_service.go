package schemasvciface

import schemamodel "github.com/sunibgram/domain/model/schema"

type ISchemaService interface {
	GetSchema() ([]schemamodel.Table, error)
}
