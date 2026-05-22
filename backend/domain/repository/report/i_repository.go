package reportrepoiface

import (
	"github.com/google/uuid"

	reportmodel "github.com/sunibgram/domain/model/report"
)

type IReportRepository interface {
	Save(report *reportmodel.Report) error
	FindAll() ([]reportmodel.Report, error)
	FindByID(id uuid.UUID) (*reportmodel.Report, error)
	UpdateStatus(id uuid.UUID, status string) error
	Delete(id uuid.UUID) error
}
