package reportrepoiface

import (
	"github.com/google/uuid"

	reportmodel "github.com/sunibgram/domain/model/report"
)

type IReportRepository interface {
	Save(report *reportmodel.Report) error
	FindAll() ([]reportmodel.Report, error)
	// FindReportedPosts returns reports targeting a thread (reported_post_id set).
	// status "" means no status filter.
	FindReportedPosts(status string) ([]reportmodel.Report, error)
	// FindReportedUsers returns reports targeting a user (reported_user_id set).
	// status "" means no status filter.
	FindReportedUsers(status string) ([]reportmodel.Report, error)
	FindByID(id uuid.UUID) (*reportmodel.Report, error)
	UpdateStatus(id uuid.UUID, status string) error
	Delete(id uuid.UUID) error
}
