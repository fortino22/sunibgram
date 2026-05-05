package repository

import (
	"errors"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"gorm.io/gorm"
)

type userFacultyRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewUserFacultyRepository(writeDB, readDB *gorm.DB) domainrepo.IUserFacultyRepository {
	return &userFacultyRepository{writeDB: writeDB, readDB: readDB}
}

func (r *userFacultyRepository) Save(uf *model.UserFaculty) error {
	return r.writeDB.Create(uf).Error
}

func (r *userFacultyRepository) FindByUserID(userID uint) (*model.UserFaculty, error) {
	var uf model.UserFaculty
	if err := r.readDB.Preload("Faculty").Where("user_id = ?", userID).First(&uf).Error; err != nil {
		return nil, err
	}
	return &uf, nil
}

func (r *userFacultyRepository) FindByFacultyID(facultyID uint, page, limit int) ([]model.UserFaculty, int64, error) {
	var rows []model.UserFaculty
	var total int64
	r.readDB.Model(&model.UserFaculty{}).Where("faculty_id = ?", facultyID).Count(&total)
	offset := (page - 1) * limit
	if err := r.readDB.Preload("User").
		Where("faculty_id = ?", facultyID).
		Offset(offset).Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *userFacultyRepository) Delete(userID, facultyID uint) error {
	result := r.writeDB.Delete(&model.UserFaculty{}, "user_id = ? AND faculty_id = ?", userID, facultyID)
	if result.RowsAffected == 0 {
		return errors.New("assignment not found")
	}
	return result.Error
}
