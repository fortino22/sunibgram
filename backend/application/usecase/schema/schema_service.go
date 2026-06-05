package schemauc

import (
	schemamodel "github.com/sunibgram/domain/model/schema"
	schemarepoiface "github.com/sunibgram/domain/repository/schema"
	schemasvciface "github.com/sunibgram/domain/service/schema"
)

type SchemaService struct {
	schemaRepo schemarepoiface.ISchemaRepository
}

func NewSchemaService(schemaRepo schemarepoiface.ISchemaRepository) schemasvciface.ISchemaService {
	return &SchemaService{schemaRepo: schemaRepo}
}

func (s *SchemaService) GetSchema() ([]schemamodel.Table, error) {
	return s.schemaRepo.GetSchema()
}
