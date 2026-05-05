package repository

import "github.com/sunibgram/domain/model"

type IFacultyRepository interface {
	Save(faculty *model.Faculty) error
	FindAll(page, limit int) ([]model.Faculty, int64, error)
	FindByID(id uint) (*model.Faculty, error)
	Update(faculty *model.Faculty) error
	Delete(id uint) error
	UpdateTotalUsers(facultyID uint, delta int) error
}
