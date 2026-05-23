package facultyrepoiface

import facultymodel "github.com/sunibgram/domain/model/faculty"

type IFacultyRepository interface {
	Save(faculty *facultymodel.Faculty) error
	FindAll(page, limit int) ([]facultymodel.Faculty, int64, error)
	FindByID(id uint) (*facultymodel.Faculty, error)
	Update(faculty *facultymodel.Faculty) error
	Delete(id uint) error
	UpdateTotalUsers(facultyID uint, delta int) error
}
