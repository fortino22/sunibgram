package reportsvciface

import (
	"github.com/google/uuid"

	reportmodel "github.com/sunibgram/domain/model/report"
)

type IReportService interface {
	Create(reporterID uint, reportedUserID, reportedPostID *uint, violationID uuid.UUID, explanation string) error
	GetAll() ([]reportmodel.ReportDetail, error)
	// GetReportedEntities returns reported threads and reported users grouped by
	// entity (each appears once with its report list). status "" = all statuses.
	GetReportedEntities(status string) (reportmodel.ReportedEntities, error)
	GetByID(id uuid.UUID) (*reportmodel.ReportDetail, error)
	UpdateStatus(id uuid.UUID, status string) error
	Delete(id uuid.UUID) error
}
