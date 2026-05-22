package threaduc

import (
	"github.com/sunibgram/application/helper"
	threadmodel "github.com/sunibgram/domain/model/thread"
)

func (s *ThreadService) GetFacultyThreads(facultyID, requesterID uint, page, limit int) ([]threadmodel.Thread, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.threadRepo.FindByFacultyID(facultyID, &requesterID, page, limit)
}
