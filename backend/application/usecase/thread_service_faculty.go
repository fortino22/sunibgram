package usecase

import (
	"github.com/sunibgram/application/helper"
	"github.com/sunibgram/domain/model"
)

func (s *ThreadService) GetFacultyThreads(facultyID uint, page, limit int) ([]model.Thread, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.threadRepo.FindByFacultyID(facultyID, page, limit)
}
