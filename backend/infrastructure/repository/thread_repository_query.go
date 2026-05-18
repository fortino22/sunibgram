package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sunibgram/domain/model"
)

func (r *threadCached) FindByID(id uint) (*model.Thread, error) {
	ctx := context.Background()
	var thread model.Thread
	key := threadKey(id)
	if err := r.cache.Get(ctx, key, &thread); err != nil {
		if err := preloadThreadDetails(r.db.readDB).Where("thread_id = ?", id).First(&thread).Error; err != nil {
			return nil, err
		}
		r.cache.Set(ctx, key, thread, r.ttl)
	}
	return &thread, nil
}

func (r *threadCached) FindRoots(requesterID *uint, page, limit int) ([]model.Thread, int64, error) {
	type cached struct {
		Threads []model.Thread
		Total   int64
	}
	ctx := context.Background()
	var result cached
	key := fmt.Sprintf("%s:viewer:%s", threadRootsKey(page, limit), threadVisibilityCachePart(requesterID))
	if err := r.cache.Get(ctx, key, &result); err != nil {
		var threads []model.Thread
		var total int64
		countQuery := applyThreadVisibility(r.db.readDB.Model(&model.Thread{}), requesterID).
			Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
			Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost)}).
			Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic)
		if err := countQuery.Count(&total).Error; err != nil {
			return nil, 0, err
		}
		offset := (page - 1) * limit
		if err := applyThreadVisibility(preloadThreadDetails(r.db.readDB), requesterID).
			Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
			Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost)}).
			Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
			Order("threads.created_at DESC").
			Offset(offset).Limit(limit).
			Find(&threads).Error; err != nil {
			return nil, 0, err
		}
		result = cached{Threads: threads, Total: total}
		if err := r.cache.Set(ctx, key, result, threadFeedTTL); err == nil {
			for _, t := range result.Threads {
				r.trackFeedCacheKey(ctx, t.ThreadID, key)
			}
		}
	}
	return result.Threads, result.Total, nil
}

func (r *threadCached) FindChain(anchorID uint) (model.Thread, []model.Thread, error) {
	var anchor model.Thread
	if err := preloadThreadParentDetails(r.db.readDB).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Where("threads.thread_id = ?", anchorID).
		Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
		First(&anchor).Error; err != nil {
		return model.Thread{}, nil, err
	}

	var children []model.Thread
	if err := preloadThreadParentDetails(r.db.readDB).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Where("threads.parent_id = ?", anchorID).
		Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
		Order("threads.created_at ASC").
		Find(&children).Error; err != nil {
		return model.Thread{}, nil, err
	}
	return anchor, children, nil
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
		if err := preloadThreadParentDetails(r.db.readDB).
			Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
			Where("threads.thread_id = ? OR threads.root_id = ?", rootID, rootID).
			Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
			Order("threads.created_at ASC").
			Offset(offset).Limit(limit).
			Find(&threads).Error; err != nil {
			return nil, 0, err
		}
		result = cached{Threads: threads, Total: total}
		if err := r.cache.Set(ctx, key, result, threadFeedTTL); err == nil {
			for _, t := range result.Threads {
				r.trackFeedCacheKey(ctx, t.ThreadID, key)
			}
		}
	}
	return result.Threads, result.Total, nil
}

func (r *threadCached) FindCommunityTop(communityID, requesterUserID uint, page, limit int) ([]model.Thread, int64, error) {
	return r.findCommunityThreads(
		communityID,
		requesterUserID,
		page,
		limit,
		"top",
		"(threads.like_count + threads.repost_count + threads.bookmark_count) DESC, threads.created_at DESC",
	)
}

func (r *threadCached) FindCommunityLatest(communityID, requesterUserID uint, page, limit int) ([]model.Thread, int64, error) {
	return r.findCommunityThreads(communityID, requesterUserID, page, limit, "latest", "threads.created_at DESC")
}

func (r *threadCached) findCommunityThreads(communityID, requesterUserID uint, page, limit int, keySuffix, orderClause string) ([]model.Thread, int64, error) {
	type cached struct {
		Threads []model.Thread
		Total   int64
	}
	ctx := context.Background()
	var result cached
	key := fmt.Sprintf("%s:viewer:%d", threadCommunityKey(communityID, page, limit, keySuffix), requesterUserID)
	if err := r.cache.Get(ctx, key, &result); err != nil {
		var threads []model.Thread
		var total int64
		requesterID := &requesterUserID
		if err := applyThreadVisibility(r.db.readDB.Model(&model.Thread{}), requesterID).
			Where("community_id = ? AND type IN ?", communityID, []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost)}).
			Count(&total).Error; err != nil {
			return nil, 0, err
		}
		offset := (page - 1) * limit
		if err := applyThreadVisibility(preloadThreadDetails(r.db.readDB), requesterID).
			Where("community_id = ? AND type IN ?", communityID, []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost)}).
			Order(orderClause).
			Offset(offset).Limit(limit).
			Find(&threads).Error; err != nil {
			return nil, 0, err
		}
		result = cached{Threads: threads, Total: total}
		if err := r.cache.Set(ctx, key, result, threadFeedTTL); err == nil {
			for _, t := range result.Threads {
				r.trackFeedCacheKey(ctx, t.ThreadID, key)
			}
		}
	}
	return result.Threads, result.Total, nil
}

func (r *threadCached) FindCommunityMedia(communityID, requesterUserID uint) ([]model.CommunityThreadMedia, error) {
	var rows []model.ThreadMedia
	if err := r.db.readDB.
		Table("thread_media").
		Select("thread_media.media_id, thread_media.thread_id, thread_media.object_name, thread_media.bucket, thread_media.created_at").
		Joins("JOIN threads ON threads.thread_id = thread_media.thread_id").
		Joins("LEFT JOIN user_settings us ON us.user_id = threads.user_id").
		Joins("LEFT JOIN user_relationships ur ON ur.follower_id = ? AND ur.following_id = threads.user_id", requesterUserID).
		Where("threads.community_id = ? AND threads.type IN ?", communityID, []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost)}).
		Where("threads.user_id = ? OR COALESCE(us.is_private, false) = false OR ur.following_id IS NOT NULL", requesterUserID).
		Order("threads.created_at DESC, thread_media.created_at ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	byThread := make(map[uint]int)
	result := make([]model.CommunityThreadMedia, 0)
	for _, media := range rows {
		idx, ok := byThread[media.ThreadID]
		if !ok {
			idx = len(result)
			byThread[media.ThreadID] = idx
			result = append(result, model.CommunityThreadMedia{ThreadID: media.ThreadID})
		}
		result[idx].Media = append(result[idx].Media, media)
	}
	return result, nil
}

func (r *threadCached) FindMyCommunityThreads(userID uint, page, limit int) ([]model.Thread, int64, error) {
	type cached struct {
		Threads []model.Thread
		Total   int64
	}
	ctx := context.Background()
	var result cached
	key := fmt.Sprintf("thread:my-communities:user:%d:page:%d:limit:%d", userID, page, limit)
	if err := r.cache.Get(ctx, key, &result); err != nil {
		var threads []model.Thread
		var total int64

		requesterID := &userID
		base := applyThreadVisibility(r.db.readDB.Model(&model.Thread{}), requesterID).
			Joins("JOIN communities ON communities.community_id = threads.community_id").
			Joins("JOIN community_members cm ON cm.community_id = threads.community_id AND cm.user_id = ?", userID).
			Where("threads.community_id IS NOT NULL").
			Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost)}).
			Where("communities.is_active = ?", true)

		if err := base.Count(&total).Error; err != nil {
			return nil, 0, err
		}

		offset := (page - 1) * limit
		if err := applyThreadVisibility(preloadThreadDetails(r.db.readDB), requesterID).
			Joins("JOIN communities ON communities.community_id = threads.community_id").
			Joins("JOIN community_members cm ON cm.community_id = threads.community_id AND cm.user_id = ?", userID).
			Where("threads.community_id IS NOT NULL").
			Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost)}).
			Where("communities.is_active = ?", true).
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

func (r *threadCached) FindAllCommunityThreads(userID uint, categoryID *uuid.UUID, page, limit int) ([]model.Thread, int64, error) {
	type cached struct {
		Threads []model.Thread
		Total   int64
	}
	ctx := context.Background()
	var result cached
	catStr := "all"
	if categoryID != nil {
		catStr = categoryID.String()
	}
	key := fmt.Sprintf("thread:all-communities:user:%d:cat:%s:page:%d:limit:%d", userID, catStr, page, limit)
	if err := r.cache.Get(ctx, key, &result); err != nil {
		var threads []model.Thread
		var total int64

		base := r.db.readDB.Model(&model.Thread{}).
			Joins("JOIN communities ON communities.community_id = threads.community_id").
			Joins("LEFT JOIN user_settings us ON us.user_id = threads.user_id").
			Joins("LEFT JOIN user_relationships ur ON ur.follower_id = ? AND ur.following_id = threads.user_id", userID).
			Where("threads.community_id IS NOT NULL").
			Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost)}).
			Where("communities.is_active = ?", true).
			Where("communities.visibility = ? OR communities.community_id IN (SELECT community_id FROM community_members WHERE user_id = ?)",
				model.CommunityVisibilityPublic, userID).
			Where("threads.user_id = ? OR COALESCE(us.is_private, false) = false OR ur.following_id IS NOT NULL", userID)

		if categoryID != nil {
			base = base.Where("communities.category_id = ?", categoryID)
		}

		if err := base.Count(&total).Error; err != nil {
			return nil, 0, err
		}

		offset := (page - 1) * limit
		query := preloadThreadDetails(r.db.readDB).
			Joins("JOIN communities ON communities.community_id = threads.community_id").
			Joins("LEFT JOIN user_settings us ON us.user_id = threads.user_id").
			Joins("LEFT JOIN user_relationships ur ON ur.follower_id = ? AND ur.following_id = threads.user_id", userID).
			Where("threads.community_id IS NOT NULL").
			Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost)}).
			Where("communities.is_active = ?", true).
			Where("communities.visibility = ? OR communities.community_id IN (SELECT community_id FROM community_members WHERE user_id = ?)",
				model.CommunityVisibilityPublic, userID).
			Where("threads.user_id = ? OR COALESCE(us.is_private, false) = false OR ur.following_id IS NOT NULL", userID)
		if categoryID != nil {
			query = query.Where("communities.category_id = ?", categoryID)
		}
		if err := query.
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

func (r *threadCached) FindAllCommunityThreadsByQueryTop(userID uint, query string, page, limit int) ([]model.Thread, int64, error) {
	return r.findAllCommunityThreadsByQuery(
		userID,
		query,
		page,
		limit,
		"top",
		"(threads.like_count + threads.repost_count + threads.bookmark_count) DESC, threads.created_at DESC",
	)
}

func (r *threadCached) FindAllCommunityThreadsByQueryLatest(userID uint, query string, page, limit int) ([]model.Thread, int64, error) {
	return r.findAllCommunityThreadsByQuery(userID, query, page, limit, "latest", "threads.created_at DESC")
}

func (r *threadCached) findAllCommunityThreadsByQuery(userID uint, query string, page, limit int, keySuffix, orderClause string) ([]model.Thread, int64, error) {
	type cached struct {
		Threads []model.Thread
		Total   int64
	}
	ctx := context.Background()
	var result cached
	query = strings.TrimSpace(query)
	queryLike := "%" + query + "%"
	key := communityThreadsQueryKey(keySuffix, strings.ToLower(query), userID, page, limit)
	if err := r.cache.Get(ctx, key, &result); err != nil {
		var threads []model.Thread
		var total int64

		base := r.db.readDB.Model(&model.Thread{}).
			Joins("JOIN communities ON communities.community_id = threads.community_id").
			Joins("LEFT JOIN user_settings us ON us.user_id = threads.user_id").
			Joins("LEFT JOIN user_relationships ur ON ur.follower_id = ? AND ur.following_id = threads.user_id", userID).
			Where("threads.community_id IS NOT NULL").
			Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost)}).
			Where("communities.is_active = ?", true).
			Where("(communities.name ILIKE ? OR threads.content ILIKE ?)", queryLike, queryLike).
			Where("communities.visibility = ? OR communities.community_id IN (SELECT community_id FROM community_members WHERE user_id = ?)",
				model.CommunityVisibilityPublic, userID).
			Where("threads.user_id = ? OR COALESCE(us.is_private, false) = false OR ur.following_id IS NOT NULL", userID)

		if err := base.Count(&total).Error; err != nil {
			return nil, 0, err
		}

		offset := (page - 1) * limit
		if err := preloadThreadDetails(r.db.readDB).
			Joins("JOIN communities ON communities.community_id = threads.community_id").
			Joins("LEFT JOIN user_settings us ON us.user_id = threads.user_id").
			Joins("LEFT JOIN user_relationships ur ON ur.follower_id = ? AND ur.following_id = threads.user_id", userID).
			Where("threads.community_id IS NOT NULL").
			Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost)}).
			Where("communities.is_active = ?", true).
			Where("(communities.name ILIKE ? OR threads.content ILIKE ?)", queryLike, queryLike).
			Where("communities.visibility = ? OR communities.community_id IN (SELECT community_id FROM community_members WHERE user_id = ?)",
				model.CommunityVisibilityPublic, userID).
			Where("threads.user_id = ? OR COALESCE(us.is_private, false) = false OR ur.following_id IS NOT NULL", userID).
			Order(orderClause).
			Offset(offset).Limit(limit).
			Find(&threads).Error; err != nil {
			return nil, 0, err
		}

		result = cached{Threads: threads, Total: total}
		if err := r.cache.Set(ctx, key, result, threadFeedTTL); err == nil {
			for _, t := range result.Threads {
				r.trackFeedCacheKey(ctx, t.ThreadID, key)
			}
		}
	}
	return result.Threads, result.Total, nil
}
