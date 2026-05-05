package repository

import "github.com/sunibgram/domain/model"

type IUserFacultyRepository interface {
	Save(uf *model.UserFaculty) error
	FindByUserID(userID uint) (*model.UserFaculty, error)
	FindByFacultyID(facultyID uint, page, limit int) ([]model.UserFaculty, int64, error)
	Delete(userID, facultyID uint) error
}
