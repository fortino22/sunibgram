package repository

import (
	"context"

	"github.com/sunibgram/domain/model"
)

func (r *threadCached) FindByID(id uint) (*model.Thread, error) {
	ctx := context.Background()
	var thread model.Thread
	key := threadKey(id)
	if err := r.cache.Get(ctx, key, &thread); err != nil {
		if err := r.db.readDB.Preload("User", preloadThreadUser).Preload("OriginalThread.User", preloadThreadUser).Preload("Community").Preload("OriginalThread.Community").Preload("User.UserFaculty.Faculty").Preload("OriginalThread.User.UserFaculty.Faculty").First(&thread, id).Error; err != nil {
			return nil, err
		}
		r.cache.Set(ctx, key, thread, r.ttl)
	}
	return &thread, nil
}

func (r *threadCached) FindRoots(page, limit int) ([]model.Thread, int64, error) {
	type cached struct {
		Threads []model.Thread
		Total   int64
	}
	ctx := context.Background()
	var result cached
	key := threadRootsKey(page, limit)
	if err := r.cache.Get(ctx, key, &result); err != nil {
		var threads []model.Thread
		var total int64
		r.db.readDB.Model(&model.Thread{}).
			Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
			Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost)}).
			Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
			Count(&total)
		offset := (page - 1) * limit
		if err := r.db.readDB.Preload("User", preloadThreadUser).Preload("OriginalThread.User", preloadThreadUser).Preload("Community").Preload("OriginalThread.Community").Preload("User.UserFaculty.Faculty").Preload("OriginalThread.User.UserFaculty.Faculty").
			Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
			Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost)}).
			Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
			Order("threads.created_at DESC").
			Offset(offset).Limit(limit).
			Find(&threads).Error; err != nil {
			return nil, 0, err
		}
		result = cached{Threads: threads, Total: total}
		r.cache.Set(ctx, key, result, threadFeedTTL)
	}
	return result.Threads, result.Total, nil
}

func (r *threadCached) FindByRootID(rootID uint, page, limit int) ([]model.Thread, int64, error) {
	type cached struct {
		Threads []model.Thread
		Total   int64
	}
	ctx := context.Background()
	var result cached
	key := threadChainKey(rootID, page, limit)
	if err := r.cache.Get(ctx, key, &result); err != nil {
		var threads []model.Thread
		var total int64
		r.db.readDB.Model(&model.Thread{}).
			Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
			Where("threads.thread_id = ? OR threads.root_id = ?", rootID, rootID).
			Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
			Count(&total)
		offset := (page - 1) * limit
		if err := r.db.readDB.Preload("User", preloadThreadUser).Preload("Community").Preload("User.UserFaculty.Faculty").
			Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
			Where("threads.thread_id = ? OR threads.root_id = ?", rootID, rootID).
			Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
			Order("threads.created_at ASC").
			Offset(offset).Limit(limit).
			Find(&threads).Error; err != nil {
			return nil, 0, err
		}
		result = cached{Threads: threads, Total: total}
		r.cache.Set(ctx, key, result, threadFeedTTL)
	}
	return result.Threads, result.Total, nil
}

func (r *threadCached) FindByCommunityID(communityID uint, page, limit int) ([]model.Thread, int64, error) {
	type cached struct {
		Threads []model.Thread
		Total   int64
	}
	ctx := context.Background()
	var result cached
	key := threadCommunityKey(communityID, page, limit)
	if err := r.cache.Get(ctx, key, &result); err != nil {
		var threads []model.Thread
		var total int64
		r.db.readDB.Model(&model.Thread{}).
			Where("community_id = ? AND type IN ?", communityID, []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost)}).
			Count(&total)
		offset := (page - 1) * limit
		if err := r.db.readDB.Preload("User", preloadThreadUser).Preload("OriginalThread.User", preloadThreadUser).Preload("User.UserFaculty.Faculty").Preload("OriginalThread.User.UserFaculty.Faculty").
			Where("community_id = ? AND type IN ?", communityID, []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost)}).
			Order("created_at DESC").
			Offset(offset).Limit(limit).
			Find(&threads).Error; err != nil {
			return nil, 0, err
		}
		result = cached{Threads: threads, Total: total}
		r.cache.Set(ctx, key, result, threadFeedTTL)
	}
	return result.Threads, result.Total, nil
}
