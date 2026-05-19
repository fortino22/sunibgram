package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"gorm.io/gorm"
)

type reportRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewReportRepository(writeDB, readDB *gorm.DB) domainrepo.IReportRepository {
	return &reportRepository{writeDB: writeDB, readDB: readDB}
}

func (r *reportRepository) Save(report *model.Report) error {
	return r.writeDB.Create(report).Error
}

func (r *reportRepository) FindAll() ([]model.Report, error) {
	var reports []model.Report
	if err := r.readDB.
		Preload("Violation").
		Order("created_at DESC").
		Find(&reports).Error; err != nil {
		return nil, err
	}
	return reports, nil
}

func (r *reportRepository) FindByID(id uuid.UUID) (*model.Report, error) {
	var report model.Report
	if err := r.readDB.
		Preload("Violation").
		First(&report, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *reportRepository) UpdateStatus(id uuid.UUID, status string) error {
	result := r.writeDB.Model(&model.Report{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("report not found")
	}
	return nil
}

func (r *reportRepository) Delete(id uuid.UUID) error {
	result := r.writeDB.Delete(&model.Report{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("report not found")
	}
	return nil
}
