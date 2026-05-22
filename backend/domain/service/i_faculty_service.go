package service

import (
	facultymodel "github.com/sunibgram/domain/model/faculty"
	usermodel "github.com/sunibgram/domain/model/user"
)

type IFacultyService interface {
	Create(name string) error
	GetAll(page, limit int) ([]facultymodel.Faculty, int64, error)
	GetByID(id uint) (*facultymodel.Faculty, error)
	Update(id uint, name string) error
	Delete(id uint) error
	SetUserFaculty(userID uint, facultyID *uint) error
	GetUserFaculty(userID uint) (*usermodel.UserFaculty, error)
}
