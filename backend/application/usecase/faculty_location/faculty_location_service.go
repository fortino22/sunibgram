package facultylocationuc

import (
	facultylocationmodel "github.com/sunibgram/domain/model/faculty_location"
	facultylocationrepoiface "github.com/sunibgram/domain/repository/faculty_location"
	facultylocationsvciface "github.com/sunibgram/domain/service/faculty_location"
)

type FacultyLocationService struct {
	repo facultylocationrepoiface.IFacultyLocationRepository
}

func NewFacultyLocationService(repo facultylocationrepoiface.IFacultyLocationRepository) facultylocationsvciface.IFacultyLocationService {
	return &FacultyLocationService{repo: repo}
}

func (s *FacultyLocationService) List(includeAll bool) ([]facultylocationmodel.FacultyLocation, error) {
	return s.repo.FindAll(includeAll)
}
