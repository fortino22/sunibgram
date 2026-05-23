package facultyuc

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/sunibgram/application/helper"
	facultymodel "github.com/sunibgram/domain/model/faculty"
	usermodel "github.com/sunibgram/domain/model/user"
	facultyrepoiface "github.com/sunibgram/domain/repository/faculty"
	userfacultyrepoiface "github.com/sunibgram/domain/repository/user_faculty"
	facultysvciface "github.com/sunibgram/domain/service/faculty"
)

type FacultyService struct {
	facultyRepo     facultyrepoiface.IFacultyRepository
	userFacultyRepo userfacultyrepoiface.IUserFacultyRepository
}

func NewFacultyService(facultyRepo facultyrepoiface.IFacultyRepository, userFacultyRepo userfacultyrepoiface.IUserFacultyRepository) facultysvciface.IFacultyService {
	return &FacultyService{facultyRepo: facultyRepo, userFacultyRepo: userFacultyRepo}
}

func (s *FacultyService) Create(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("faculty name is required")
	}
	if utf8.RuneCountInString(name) > 100 {
		return errors.New("facultyName must be 100 characters or less")
	}
	return s.facultyRepo.Save(&facultymodel.Faculty{FacultyName: name})
}

func (s *FacultyService) GetAll(page, limit int) ([]facultymodel.Faculty, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 10)
	return s.facultyRepo.FindAll(page, limit)
}

func (s *FacultyService) GetByID(id uint) (*facultymodel.Faculty, error) {
	f, err := s.facultyRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("faculty not found")
	}
	return f, nil
}

func (s *FacultyService) Update(id uint, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("faculty name is required")
	}
	if utf8.RuneCountInString(name) > 100 {
		return errors.New("facultyName must be 100 characters or less")
	}
	f, err := s.facultyRepo.FindByID(id)
	if err != nil {
		return errors.New("faculty not found")
	}
	f.FacultyName = name
	return s.facultyRepo.Update(f)
}

func (s *FacultyService) Delete(id uint) error {
	if _, err := s.facultyRepo.FindByID(id); err != nil {
		return errors.New("faculty not found")
	}
	return s.facultyRepo.Delete(id)
}

func (s *FacultyService) GetUserFaculty(userID uint) (*usermodel.UserFaculty, error) {
	uf, err := s.userFacultyRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("faculty not found")
	}
	return uf, nil
}

func (s *FacultyService) SetUserFaculty(userID uint, facultyID *uint) error {
	existing, _ := s.userFacultyRepo.FindByUserID(userID)

	if facultyID == nil {
		if existing == nil {
			return nil
		}
		if err := s.userFacultyRepo.Delete(userID, existing.FacultyID); err != nil {
			return err
		}
		return s.facultyRepo.UpdateTotalUsers(existing.FacultyID, -1)
	}

	if _, err := s.facultyRepo.FindByID(*facultyID); err != nil {
		return errors.New("faculty not found")
	}

	if existing != nil {
		if existing.FacultyID == *facultyID {
			return nil
		}
		if err := s.userFacultyRepo.Delete(userID, existing.FacultyID); err != nil {
			return err
		}
		_ = s.facultyRepo.UpdateTotalUsers(existing.FacultyID, -1)
	}

	if err := s.userFacultyRepo.Save(&usermodel.UserFaculty{UserID: userID, FacultyID: *facultyID}); err != nil {
		return err
	}
	return s.facultyRepo.UpdateTotalUsers(*facultyID, 1)
}
