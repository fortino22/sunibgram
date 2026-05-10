package repository

import "github.com/sunibgram/domain/model"

func (r *threadCached) findUserFeed(requesterID, userID uint, types []string, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	var total int64
	r.db.readDB.Model(&model.Thread{}).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN community_members ON community_members.community_id = threads.community_id AND community_members.user_id = ?", requesterID).
		Where("threads.user_id = ? AND threads.type IN ?", userID, types).
		Where("threads.community_id IS NULL OR communities.visibility = ? OR community_members.user_id = ?", model.CommunityVisibilityPublic, requesterID).
		Count(&total)

	offset := (page - 1) * limit
	var threads []model.Thread
	if err := r.db.readDB.Preload("User", preloadThreadUser).Preload("OriginalThread.User", preloadThreadUser).Preload("Community").Preload("OriginalThread.Community").Preload("User.UserFaculty.Faculty").Preload("OriginalThread.User.UserFaculty.Faculty").
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN community_members ON community_members.community_id = threads.community_id AND community_members.user_id = ?", requesterID).
		Where("threads.user_id = ? AND threads.type IN ?", userID, types).
		Where("threads.community_id IS NULL OR communities.visibility = ? OR community_members.user_id = ?", model.CommunityVisibilityPublic, requesterID).
		Order("threads.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&threads).Error; err != nil {
		return nil, nil, 0, err
	}

	parentIDSet := make(map[uint]struct{})
	for _, t := range threads {
		if t.ParentID != nil {
			parentIDSet[*t.ParentID] = struct{}{}
		}
	}
	parentMap := make(map[uint]*model.Thread)
	if len(parentIDSet) > 0 {
		pids := make([]uint, 0, len(parentIDSet))
		for id := range parentIDSet {
			pids = append(pids, id)
		}
		var parentThreads []model.Thread
		if err := r.db.readDB.Preload("User", preloadThreadUser).Preload("Community").Preload("User.UserFaculty.Faculty").Where("thread_id IN ?", pids).Find(&parentThreads).Error; err != nil {
			return nil, nil, 0, err
		}
		for i := range parentThreads {
			parentMap[parentThreads[i].ThreadID] = &parentThreads[i]
		}
	}

	parents := make([]*model.Thread, len(threads))
	for i, t := range threads {
		if t.ParentID != nil {
			parents[i] = parentMap[*t.ParentID]
		}
	}
	return threads, parents, total, nil
}

func (r *threadCached) FindPostsByUserID(requesterID, userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	return r.findUserFeed(requesterID, userID, []string{
		string(model.ThreadTypePost),
		string(model.ThreadTypeRepost),
	}, page, limit)
}

func (r *threadCached) FindReplyTabByUserID(requesterID, userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	return r.findUserFeed(requesterID, userID, []string{
		string(model.ThreadTypePost),
		string(model.ThreadTypeReply),
		string(model.ThreadTypeRepost),
	}, page, limit)
}

func (r *threadCached) FindMediaByUserID(requesterID, userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	types := []string{
		string(model.ThreadTypePost),
		string(model.ThreadTypeReply),
	}

	base := r.db.readDB.Model(&model.Thread{}).
		Joins("JOIN thread_media ON thread_media.thread_id = threads.thread_id").
		Joins("LEFT JOIN community_members ON community_members.community_id = threads.community_id AND community_members.user_id = ?", requesterID).
		Where("threads.user_id = ? AND threads.type IN ?", userID, types).
		Where("threads.community_id IS NULL OR community_members.user_id = ?", requesterID)

	var total int64
	if err := base.Distinct("threads.thread_id").Count(&total).Error; err != nil {
		return nil, nil, 0, err
	}

	offset := (page - 1) * limit
	var ids []uint
	if err := base.Select("threads.thread_id").
		Group("threads.thread_id").
		Order("MAX(threads.created_at) DESC").
		Offset(offset).Limit(limit).
		Pluck("threads.thread_id", &ids).Error; err != nil {
		return nil, nil, 0, err
	}
	if len(ids) == 0 {
		return []model.Thread{}, []*model.Thread{}, total, nil
	}

	var threads []model.Thread
	if err := r.db.readDB.Preload("User", preloadThreadUser).Preload("OriginalThread.User", preloadThreadUser).Preload("Community").Preload("OriginalThread.Community").Preload("User.UserFaculty.Faculty").Preload("OriginalThread.User.UserFaculty.Faculty").
		Where("thread_id IN ?", ids).
		Order("created_at DESC").
		Find(&threads).Error; err != nil {
		return nil, nil, 0, err
	}

	parentIDSet := make(map[uint]struct{})
	for _, t := range threads {
		if t.ParentID != nil {
			parentIDSet[*t.ParentID] = struct{}{}
		}
	}
	parentMap := make(map[uint]*model.Thread)
	if len(parentIDSet) > 0 {
		pids := make([]uint, 0, len(parentIDSet))
		for id := range parentIDSet {
			pids = append(pids, id)
		}
		var parentThreads []model.Thread
		if err := r.db.readDB.Preload("User", preloadThreadUser).Preload("Community").Preload("User.UserFaculty.Faculty").Where("thread_id IN ?", pids).Find(&parentThreads).Error; err != nil {
			return nil, nil, 0, err
		}
		for i := range parentThreads {
			parentMap[parentThreads[i].ThreadID] = &parentThreads[i]
		}
	}

	parents := make([]*model.Thread, len(threads))
	for i, t := range threads {
		if t.ParentID != nil {
			parents[i] = parentMap[*t.ParentID]
		}
	}
	return threads, parents, total, nil
}

func (r *threadCached) FindFollowingFeed(userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	types := []string{
		string(model.ThreadTypePost),
		string(model.ThreadTypeRepost),
		string(model.ThreadTypeReply),
	}

	var total int64
	r.db.readDB.Model(&model.Thread{}).
		Joins("JOIN user_relationships ON user_relationships.following_id = threads.user_id").
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN community_members ON community_members.community_id = threads.community_id AND community_members.user_id = ?", userID).
		Where("user_relationships.follower_id = ? AND threads.type IN ?", userID, types).
		Where("threads.community_id IS NULL OR communities.visibility = ? OR community_members.user_id = ?",
			model.CommunityVisibilityPublic, userID).
		Count(&total)

	offset := (page - 1) * limit
	var threads []model.Thread
	if err := r.db.readDB.
		Preload("User", preloadThreadUser).Preload("OriginalThread.User", preloadThreadUser).Preload("Community").Preload("OriginalThread.Community").Preload("User.UserFaculty.Faculty").Preload("OriginalThread.User.UserFaculty.Faculty").
		Joins("JOIN user_relationships ON user_relationships.following_id = threads.user_id").
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN community_members ON community_members.community_id = threads.community_id AND community_members.user_id = ?", userID).
		Where("user_relationships.follower_id = ? AND threads.type IN ?", userID, types).
		Where("threads.community_id IS NULL OR communities.visibility = ? OR community_members.user_id = ?",
			model.CommunityVisibilityPublic, userID).
		Order("threads.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&threads).Error; err != nil {
		return nil, nil, 0, err
	}

	parentIDSet := make(map[uint]struct{})
	for _, t := range threads {
		if t.ParentID != nil {
			parentIDSet[*t.ParentID] = struct{}{}
		}
	}
	parentMap := make(map[uint]*model.Thread)
	if len(parentIDSet) > 0 {
		pids := make([]uint, 0, len(parentIDSet))
		for id := range parentIDSet {
			pids = append(pids, id)
		}
		var parentThreads []model.Thread
		if err := r.db.readDB.Preload("User", preloadThreadUser).Preload("Community").Preload("User.UserFaculty.Faculty").Where("thread_id IN ?", pids).Find(&parentThreads).Error; err != nil {
			return nil, nil, 0, err
		}
		for i := range parentThreads {
			parentMap[parentThreads[i].ThreadID] = &parentThreads[i]
		}
	}

	parents := make([]*model.Thread, len(threads))
	for i, t := range threads {
		if t.ParentID != nil {
			parents[i] = parentMap[*t.ParentID]
		}
	}
	return threads, parents, total, nil
}
