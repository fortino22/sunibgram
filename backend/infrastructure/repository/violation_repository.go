package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	violationmodel "github.com/sunibgram/domain/model/violation"
	domainrepo "github.com/sunibgram/domain/repository"
)

type violationRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewViolationRepository(writeDB, readDB *gorm.DB) domainrepo.IViolationRepository {
	return &violationRepository{writeDB: writeDB, readDB: readDB}
}

func (r *violationRepository) Save(violation *violationmodel.Violation) error {
	return r.writeDB.Create(violation).Error
}

func (r *violationRepository) FindAll() ([]violationmodel.Violation, error) {
	var violations []violationmodel.Violation
	if err := r.readDB.Order("name ASC").Find(&violations).Error; err != nil {
		return nil, err
	}
	return violations, nil
}

func (r *violationRepository) FindByID(id uuid.UUID) (*violationmodel.Violation, error) {
	var violation violationmodel.Violation
	if err := r.readDB.First(&violation, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &violation, nil
}

func (r *violationRepository) Update(violation *violationmodel.Violation) error {
	result := r.writeDB.Model(&violationmodel.Violation{}).
		Where("id = ?", violation.ID).
		Updates(map[string]any{
			"name":        violation.Name,
			"description": violation.Description,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("violation not found")
	}
	return nil
}

func (r *violationRepository) Delete(id uuid.UUID) error {
	result := r.writeDB.Delete(&violationmodel.Violation{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("violation not found")
	}
	return nil
}
