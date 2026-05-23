package threadrepo

import (
	communitymodel "github.com/sunibgram/domain/model/community"
	threadmodel "github.com/sunibgram/domain/model/thread"
)

func (r *threadCached) FindByFacultyID(facultyID uint, requesterID *uint, page, limit int) ([]threadmodel.Thread, int64, error) {
	var total int64
	base := ApplyThreadVisibility(r.db.readDB.Model(&threadmodel.Thread{}), requesterID).
		Joins("JOIN user_faculties ON user_faculties.user_id = threads.user_id").
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Where("user_faculties.faculty_id = ?", facultyID).
		Where("threads.type IN ?", []string{string(threadmodel.ThreadTypePost), string(threadmodel.ThreadTypeRepost)}).
		Where("threads.community_id IS NULL OR communities.visibility = ?", communitymodel.CommunityVisibilityPublic)

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var threads []threadmodel.Thread
	offset := (page - 1) * limit
	if err := ApplyThreadVisibility(PreloadThreadDetails(r.db.readDB), requesterID).
		Joins("JOIN user_faculties ON user_faculties.user_id = threads.user_id").
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Where("user_faculties.faculty_id = ?", facultyID).
		Where("threads.type IN ?", []string{string(threadmodel.ThreadTypePost), string(threadmodel.ThreadTypeRepost)}).
		Where("threads.community_id IS NULL OR communities.visibility = ?", communitymodel.CommunityVisibilityPublic).
		Order("threads.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&threads).Error; err != nil {
		return nil, 0, err
	}

	return threads, total, nil
}
