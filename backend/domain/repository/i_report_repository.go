package repository

import (
	"github.com/google/uuid"

	"github.com/sunibgram/domain/model"
)

type IReportRepository interface {
	Save(report *model.Report) error
	FindAll() ([]model.Report, error)
	FindByID(id uuid.UUID) (*model.Report, error)
	UpdateStatus(id uuid.UUID, status string) error
	Delete(id uuid.UUID) error
}
