package repository

import (
	"github.com/sunibgram/domain/model"
)

func (r *threadCached) FindByFacultyID(facultyID uint, page, limit int) ([]model.Thread, int64, error) {
	var total int64
	base := r.db.readDB.Model(&model.Thread{}).
		Joins("JOIN user_faculties ON user_faculties.user_id = threads.user_id").
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Where("user_faculties.faculty_id = ?", facultyID).
		Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost)}).
		Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic)

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var threads []model.Thread
	offset := (page - 1) * limit
	if err := preloadThreadDetails(r.db.readDB).
		Joins("JOIN user_faculties ON user_faculties.user_id = threads.user_id").
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Where("user_faculties.faculty_id = ?", facultyID).
		Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost)}).
		Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
		Order("threads.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&threads).Error; err != nil {
		return nil, 0, err
	}

	return threads, total, nil
}
