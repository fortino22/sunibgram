package reportrepo

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	reportmodel "github.com/sunibgram/domain/model/report"
	reportrepoiface "github.com/sunibgram/domain/repository/report"
)

type reportRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewReportRepository(writeDB, readDB *gorm.DB) reportrepoiface.IReportRepository {
	return &reportRepository{writeDB: writeDB, readDB: readDB}
}

func (r *reportRepository) Save(report *reportmodel.Report) error {
	return r.writeDB.Create(report).Error
}

func (r *reportRepository) FindAll() ([]reportmodel.Report, error) {
	var reports []reportmodel.Report
	if err := r.readDB.
		Preload("Violation").
		Order("created_at DESC").
		Find(&reports).Error; err != nil {
		return nil, err
	}
	return reports, nil
}

func (r *reportRepository) FindReportedPosts(status string) ([]reportmodel.Report, error) {
	var reports []reportmodel.Report
	q := r.readDB.Preload("Violation").Where("reported_post_id IS NOT NULL")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Order("created_at DESC").Find(&reports).Error; err != nil {
		return nil, err
	}
	return reports, nil
}

func (r *reportRepository) FindReportedUsers(status string) ([]reportmodel.Report, error) {
	var reports []reportmodel.Report
	q := r.readDB.Preload("Violation").Where("reported_user_id IS NOT NULL")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Order("created_at DESC").Find(&reports).Error; err != nil {
		return nil, err
	}
	return reports, nil
}

func (r *reportRepository) FindByID(id uuid.UUID) (*reportmodel.Report, error) {
	var report reportmodel.Report
	if err := r.readDB.
		Preload("Violation").
		First(&report, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *reportRepository) UpdateStatus(id uuid.UUID, status string) error {
	result := r.writeDB.Model(&reportmodel.Report{}).
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
	result := r.writeDB.Delete(&reportmodel.Report{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("report not found")
	}
	return nil
}
