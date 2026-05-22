package service

import (
	"github.com/google/uuid"

	"github.com/sunibgram/domain/model"
)

type IReportService interface {
	Create(reporterID uint, reportedUserID, reportedPostID *uint, violationID uuid.UUID, explanation string) error
	GetAll() ([]model.ReportDetail, error)
	GetByID(id uuid.UUID) (*model.ReportDetail, error)
	UpdateStatus(id uuid.UUID, status string) error
	Delete(id uuid.UUID) error
}
