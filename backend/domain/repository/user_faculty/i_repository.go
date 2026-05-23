package userfacultyrepoiface

import usermodel "github.com/sunibgram/domain/model/user"

type IUserFacultyRepository interface {
	Save(uf *usermodel.UserFaculty) error
	FindByUserID(userID uint) (*usermodel.UserFaculty, error)
	FindByUserIDs(ids []uint) ([]usermodel.UserFaculty, error)
	FindByFacultyID(facultyID uint, page, limit int) ([]usermodel.UserFaculty, int64, error)
	Delete(userID, facultyID uint) error
}
