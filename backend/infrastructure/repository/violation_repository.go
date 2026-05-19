package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"gorm.io/gorm"
)

type violationRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewViolationRepository(writeDB, readDB *gorm.DB) domainrepo.IViolationRepository {
	return &violationRepository{writeDB: writeDB, readDB: readDB}
}

func (r *violationRepository) Save(violation *model.Violation) error {
	return r.writeDB.Create(violation).Error
}

func (r *violationRepository) FindAll() ([]model.Violation, error) {
	var violations []model.Violation
	if err := r.readDB.Order("name ASC").Find(&violations).Error; err != nil {
		return nil, err
	}
	return violations, nil
}

func (r *violationRepository) FindByID(id uuid.UUID) (*model.Violation, error) {
	var violation model.Violation
	if err := r.readDB.First(&violation, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &violation, nil
}

func (r *violationRepository) Update(violation *model.Violation) error {
	result := r.writeDB.Model(&model.Violation{}).
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
	result := r.writeDB.Delete(&model.Violation{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("violation not found")
	}
	return nil
}
