package repository

import (
	"errors"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"gorm.io/gorm"
)

type facultyRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewFacultyRepository(writeDB, readDB *gorm.DB) domainrepo.IFacultyRepository {
	return &facultyRepository{writeDB: writeDB, readDB: readDB}
}

func (r *facultyRepository) Save(faculty *model.Faculty) error {
	return r.writeDB.Create(faculty).Error
}

func (r *facultyRepository) FindAll(page, limit int) ([]model.Faculty, int64, error) {
	var faculties []model.Faculty
	var total int64
	r.readDB.Model(&model.Faculty{}).Count(&total)
	offset := (page - 1) * limit
	if err := r.readDB.Offset(offset).Limit(limit).Find(&faculties).Error; err != nil {
		return nil, 0, err
	}
	return faculties, total, nil
}

func (r *facultyRepository) FindByID(id uint) (*model.Faculty, error) {
	var faculty model.Faculty
	if err := r.readDB.First(&faculty, "faculty_id = ?", id).Error; err != nil {
		return nil, err
	}
	return &faculty, nil
}

func (r *facultyRepository) Update(faculty *model.Faculty) error {
	result := r.writeDB.Model(faculty).Where("faculty_id = ?", faculty.FacultyID).
		UpdateColumn("faculty_name", faculty.FacultyName)
	if result.RowsAffected == 0 {
		return errors.New("faculty not found")
	}
	return result.Error
}

func (r *facultyRepository) Delete(id uint) error {
	result := r.writeDB.Delete(&model.Faculty{}, "faculty_id = ?", id)
	if result.RowsAffected == 0 {
		return errors.New("faculty not found")
	}
	return result.Error
}

func (r *facultyRepository) UpdateTotalUsers(facultyID uint, delta int) error {
	var expr interface{}
	if delta >= 0 {
		expr = gorm.Expr("total_users + ?", delta)
	} else {
		expr = gorm.Expr("GREATEST(total_users + ?, 0)", delta)
	}
	return r.writeDB.Model(&model.Faculty{}).
		Where("faculty_id = ?", facultyID).
		UpdateColumn("total_users", expr).Error
}
