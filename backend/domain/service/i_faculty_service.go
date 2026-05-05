package service

import "github.com/sunibgram/domain/model"

type IFacultyService interface {
	Create(name string) error
	GetAll(page, limit int) ([]model.Faculty, int64, error)
	GetByID(id uint) (*model.Faculty, error)
	Update(id uint, name string) error
	Delete(id uint) error
	SetUserFaculty(userID uint, facultyID *uint) error
}
