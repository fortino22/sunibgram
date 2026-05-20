package repository

import (
	"context"
	"strings"
	"time"

	"github.com/sunibgram/domain/model"
	"gorm.io/gorm"
)

const followingFeedTTL = 30 * time.Second

func (r *threadCached) findUserFeed(requesterID, userID uint, types []string, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	ctx := context.Background()
	tab := strings.Join(types, ",")
	key := threadUserFeedKey(userID, tab, requesterID, page, limit)
	type cachedFeed struct {
		Threads []model.Thread
		Parents []*model.Thread
		Total   int64
	}
	var cached cachedFeed
	if err := r.cache.Get(ctx, key, &cached); err == nil {
		return cached.Threads, cached.Parents, cached.Total, nil
	}

	var total int64
	requester := &requesterID
	if err := applyThreadVisibility(r.db.readDB.Model(&model.Thread{}), requester).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN community_members ON community_members.community_id = threads.community_id AND community_members.user_id = ?", requesterID).
		Where("threads.user_id = ? AND threads.type IN ?", userID, types).
		Where("threads.community_id IS NULL OR communities.visibility = ? OR community_members.user_id = ?", model.CommunityVisibilityPublic, requesterID).
		Count(&total).Error; err != nil {
		return nil, nil, 0, err
	}

	offset := (page - 1) * limit
	var threads []model.Thread
	if err := applyThreadVisibility(preloadThreadDetails(r.db.readDB), requester).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN community_members ON community_members.community_id = threads.community_id AND community_members.user_id = ?", requesterID).
		Where("threads.user_id = ? AND threads.type IN ?", userID, types).
		Where("threads.community_id IS NULL OR communities.visibility = ? OR community_members.user_id = ?", model.CommunityVisibilityPublic, requesterID).
		Order("threads.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&threads).Error; err != nil {
		return nil, nil, 0, err
	}

	threads, parents, err := r.attachVisibleParents(threads, requester)
	if err != nil {
		return nil, nil, 0, err
	}
	if err := r.cache.Set(ctx, key, cachedFeed{Threads: threads, Parents: parents, Total: total}, threadFeedTTL); err == nil {
		for _, t := range threads {
			r.trackFeedCacheKey(ctx, t.ThreadID, key)
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
		string(model.ThreadTypeReply),
	}, page, limit)
}

func (r *threadCached) FindMediaByUserID(requesterID, userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	ctx := context.Background()
	key := threadMediaFeedKey(userID, requesterID, page, limit)
	type cachedMedia struct {
		Threads []model.Thread
		Parents []*model.Thread
		Total   int64
	}
	var cached cachedMedia
	if err := r.cache.Get(ctx, key, &cached); err == nil {
		return cached.Threads, cached.Parents, cached.Total, nil
	}

	types := []string{
		string(model.ThreadTypePost),
		string(model.ThreadTypeReply),
	}

	requester := &requesterID
	baseJoins := func() *gorm.DB {
		return applyThreadVisibility(r.db.readDB.Model(&model.Thread{}), requester).
			Joins("JOIN thread_media ON thread_media.thread_id = threads.thread_id").
			Joins("LEFT JOIN community_members ON community_members.community_id = threads.community_id AND community_members.user_id = ?", requesterID).
			Where("threads.user_id = ? AND threads.type IN ?", userID, types).
			Where("threads.community_id IS NULL OR community_members.user_id = ?", requesterID)
	}

	var total int64
	if err := baseJoins().Distinct("threads.thread_id").Count(&total).Error; err != nil {
		return nil, nil, 0, err
	}

	offset := (page - 1) * limit
	var ids []uint
	if err := baseJoins().Select("threads.thread_id").
		Group("threads.thread_id").
		Order("MAX(threads.created_at) DESC").
		Offset(offset).Limit(limit).
		Pluck("threads.thread_id", &ids).Error; err != nil {
		return nil, nil, 0, err
	}
	if len(ids) == 0 {
		r.cache.Set(ctx, key, cachedMedia{Threads: []model.Thread{}, Parents: []*model.Thread{}, Total: total}, threadFeedTTL)
		return []model.Thread{}, []*model.Thread{}, total, nil
	}

	var threads []model.Thread
	if err := preloadThreadDetails(r.db.readDB).
		Where("thread_id IN ?", ids).
		Order("created_at DESC").
		Find(&threads).Error; err != nil {
		return nil, nil, 0, err
	}

	threads, parents, err := r.attachVisibleParents(threads, requester)
	if err != nil {
		return nil, nil, 0, err
	}
	if err := r.cache.Set(ctx, key, cachedMedia{Threads: threads, Parents: parents, Total: total}, threadFeedTTL); err == nil {
		for _, t := range threads {
			r.trackFeedCacheKey(ctx, t.ThreadID, key)
		}
	}
	return threads, parents, total, nil
}

func (r *threadCached) FindFollowingFeed(userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	ctx := context.Background()
	key := threadFollowingFeedKey(userID, page, limit)
	type cachedFollowing struct {
		Threads []model.Thread
		Parents []*model.Thread
		Total   int64
	}
	var cached cachedFollowing
	if err := r.cache.Get(ctx, key, &cached); err == nil {
		return cached.Threads, cached.Parents, cached.Total, nil
	}

	types := []string{
		string(model.ThreadTypePost),
		string(model.ThreadTypeRepost),
		string(model.ThreadTypeReply),
	}

	var total int64
	requester := &userID
	if err := applyThreadVisibility(r.db.readDB.Model(&model.Thread{}), requester).
		Joins("JOIN user_relationships ON user_relationships.following_id = threads.user_id").
		Joins("LEFT JOIN threads parent_threads ON parent_threads.thread_id = threads.parent_id").
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN community_members ON community_members.community_id = threads.community_id AND community_members.user_id = ?", userID).
		Where("user_relationships.follower_id = ? AND threads.type IN ?", userID, types).
		Where("threads.type <> ? OR parent_threads.user_id <> threads.user_id", model.ThreadTypeReply).
		Where("threads.community_id IS NULL OR communities.visibility = ? OR community_members.user_id = ?",
			model.CommunityVisibilityPublic, userID).
		Count(&total).Error; err != nil {
		return nil, nil, 0, err
	}

	offset := (page - 1) * limit
	var threads []model.Thread
	if err := applyThreadVisibility(preloadThreadDetails(r.db.readDB), requester).
		Joins("JOIN user_relationships ON user_relationships.following_id = threads.user_id").
		Joins("LEFT JOIN threads parent_threads ON parent_threads.thread_id = threads.parent_id").
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN community_members ON community_members.community_id = threads.community_id AND community_members.user_id = ?", userID).
		Where("user_relationships.follower_id = ? AND threads.type IN ?", userID, types).
		Where("threads.type <> ? OR parent_threads.user_id <> threads.user_id", model.ThreadTypeReply).
		Where("threads.community_id IS NULL OR communities.visibility = ? OR community_members.user_id = ?",
			model.CommunityVisibilityPublic, userID).
		Order("threads.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&threads).Error; err != nil {
		return nil, nil, 0, err
	}

	threads, parents, err := r.attachVisibleParents(threads, requester)
	if err != nil {
		return nil, nil, 0, err
	}
	if err := r.cache.Set(ctx, key, cachedFollowing{Threads: threads, Parents: parents, Total: total}, followingFeedTTL); err == nil {
		for _, t := range threads {
			r.trackFeedCacheKey(ctx, t.ThreadID, key)
		}
	}
	return threads, parents, total, nil
}

func (r *threadCached) attachVisibleParents(threads []model.Thread, requesterID *uint) ([]model.Thread, []*model.Thread, error) {
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
		if err := applyThreadVisibility(preloadThreadParentDetails(r.db.readDB), requesterID).
			Where("threads.thread_id IN ?", pids).
			Find(&parentThreads).Error; err != nil {
			return nil, nil, err
		}
		for i := range parentThreads {
			parentMap[parentThreads[i].ThreadID] = &parentThreads[i]
		}
	}

	outThreads := make([]model.Thread, 0, len(threads))
	parents := make([]*model.Thread, 0, len(threads))
	for _, thread := range threads {
		var parent *model.Thread
		if thread.ParentID != nil {
			var ok bool
			parent, ok = parentMap[*thread.ParentID]
			if !ok {
				continue
			}
		}
		outThreads = append(outThreads, thread)
		parents = append(parents, parent)
	}
	return outThreads, parents, nil
}
