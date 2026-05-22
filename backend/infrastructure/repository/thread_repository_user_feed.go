package repository

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"

	communitymodel "github.com/sunibgram/domain/model/community"
	threadmodel "github.com/sunibgram/domain/model/thread"
)

const followingFeedTTL = 30 * time.Second

func (r *threadCached) findUserFeed(requesterID, userID uint, types []string, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	ctx := context.Background()
	tab := strings.Join(types, ",")
	key := threadUserFeedKey(userID, tab, requesterID, page, limit)
	type cachedFeed struct {
		Threads []threadmodel.Thread
		Parents []*threadmodel.Thread
		Total   int64
	}
	var cached cachedFeed
	if err := r.cache.Get(ctx, key, &cached); err == nil {
		return cached.Threads, cached.Parents, cached.Total, nil
	}

	var total int64
	requester := &requesterID
	if err := applyThreadVisibility(r.db.readDB.Model(&threadmodel.Thread{}), requester).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN community_members ON community_members.community_id = threads.community_id AND community_members.user_id = ?", requesterID).
		Where("threads.user_id = ? AND threads.type IN ?", userID, types).
		Where("threads.community_id IS NULL OR communities.visibility = ? OR community_members.user_id = ?", communitymodel.CommunityVisibilityPublic, requesterID).
		Count(&total).Error; err != nil {
		return nil, nil, 0, err
	}

	offset := (page - 1) * limit
	var threads []threadmodel.Thread
	if err := applyThreadVisibility(preloadThreadDetails(r.db.readDB), requester).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN community_members ON community_members.community_id = threads.community_id AND community_members.user_id = ?", requesterID).
		Where("threads.user_id = ? AND threads.type IN ?", userID, types).
		Where("threads.community_id IS NULL OR communities.visibility = ? OR community_members.user_id = ?", communitymodel.CommunityVisibilityPublic, requesterID).
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
		r.trackUserFeedKey(ctx, userID, key)
		if requesterID != 0 {
			r.cache.RegisterCacheKey(ctx, viewerThreadFeedsTag(requesterID), key)
		}
		for _, t := range threads {
			r.trackFeedCacheKey(ctx, t.ThreadID, key)
		}
	}
	return threads, parents, total, nil
}

func (r *threadCached) FindPostsByUserID(requesterID, userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	return r.findUserFeed(requesterID, userID, []string{
		string(threadmodel.ThreadTypePost),
		string(threadmodel.ThreadTypeRepost),
	}, page, limit)
}

func (r *threadCached) FindReplyTabByUserID(requesterID, userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	return r.findUserFeed(requesterID, userID, []string{
		string(threadmodel.ThreadTypeReply),
	}, page, limit)
}

func (r *threadCached) FindMediaByUserID(requesterID, userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	ctx := context.Background()
	key := threadMediaFeedKey(userID, requesterID, page, limit)
	type cachedMedia struct {
		Threads []threadmodel.Thread
		Parents []*threadmodel.Thread
		Total   int64
	}
	var cached cachedMedia
	if err := r.cache.Get(ctx, key, &cached); err == nil {
		return cached.Threads, cached.Parents, cached.Total, nil
	}

	types := []string{
		string(threadmodel.ThreadTypePost),
		string(threadmodel.ThreadTypeReply),
	}

	requester := &requesterID
	baseJoins := func() *gorm.DB {
		return applyThreadVisibility(r.db.readDB.Model(&threadmodel.Thread{}), requester).
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
		if err := r.cache.Set(ctx, key, cachedMedia{Threads: []threadmodel.Thread{}, Parents: []*threadmodel.Thread{}, Total: total}, threadFeedTTL); err == nil {
			r.trackUserFeedKey(ctx, userID, key)
			if requesterID != 0 {
				r.cache.RegisterCacheKey(ctx, viewerThreadFeedsTag(requesterID), key)
			}
		}
		return []threadmodel.Thread{}, []*threadmodel.Thread{}, total, nil
	}

	var threads []threadmodel.Thread
	if err := preloadThreadDetails(r.db.readDB).
		Where("threads.thread_id IN ?", ids).
		Order("threads.created_at DESC").
		Find(&threads).Error; err != nil {
		return nil, nil, 0, err
	}

	threads, parents, err := r.attachVisibleParents(threads, requester)
	if err != nil {
		return nil, nil, 0, err
	}
	if err := r.cache.Set(ctx, key, cachedMedia{Threads: threads, Parents: parents, Total: total}, threadFeedTTL); err == nil {
		r.trackUserFeedKey(ctx, userID, key)
		if requesterID != 0 {
			r.cache.RegisterCacheKey(ctx, viewerThreadFeedsTag(requesterID), key)
		}
		for _, t := range threads {
			r.trackFeedCacheKey(ctx, t.ThreadID, key)
		}
	}
	return threads, parents, total, nil
}

func (r *threadCached) FindFollowingFeed(userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	ctx := context.Background()
	key := threadFollowingFeedKey(userID, page, limit)
	type cachedFollowing struct {
		Threads []threadmodel.Thread
		Parents []*threadmodel.Thread
		Total   int64
	}
	var cached cachedFollowing
	if err := r.cache.Get(ctx, key, &cached); err == nil {
		return cached.Threads, cached.Parents, cached.Total, nil
	}

	types := []string{
		string(threadmodel.ThreadTypePost),
		string(threadmodel.ThreadTypeRepost),
		string(threadmodel.ThreadTypeReply),
	}

	var total int64
	requester := &userID
	countQuery := applyThreadVisibility(r.db.readDB.Model(&threadmodel.Thread{}), requester).
		Joins("JOIN user_relationships ON user_relationships.following_id = threads.user_id").
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN community_members ON community_members.community_id = threads.community_id AND community_members.user_id = ?", userID).
		Where("user_relationships.follower_id = ? AND threads.type IN ?", userID, types).
		Where("threads.community_id IS NULL OR communities.visibility = ? OR community_members.user_id = ?",
			communitymodel.CommunityVisibilityPublic, userID)
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, nil, 0, err
	}

	offset := (page - 1) * limit
	var threads []threadmodel.Thread
	query := applyThreadVisibility(preloadThreadDetails(r.db.readDB), requester).
		Joins("JOIN user_relationships ON user_relationships.following_id = threads.user_id").
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN community_members ON community_members.community_id = threads.community_id AND community_members.user_id = ?", userID).
		Where("user_relationships.follower_id = ? AND threads.type IN ?", userID, types).
		Where("threads.community_id IS NULL OR communities.visibility = ? OR community_members.user_id = ?",
			communitymodel.CommunityVisibilityPublic, userID)
	if err := query.
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
		r.trackUserFeedKey(ctx, userID, key)
		r.cache.RegisterCacheKey(ctx, globalThreadFeedsTag(), key)
		r.cache.RegisterCacheKey(ctx, threadFollowingGlobalTag(), key)
		for _, t := range threads {
			r.trackFeedCacheKey(ctx, t.ThreadID, key)
		}
	}
	return threads, parents, total, nil
}

func (r *threadCached) AttachReplyParents(threads []threadmodel.Thread, requesterID *uint) ([]threadmodel.Thread, []*threadmodel.Thread, error) {
	return r.attachVisibleParents(threads, requesterID)
}

func (r *threadCached) attachVisibleParents(threads []threadmodel.Thread, requesterID *uint) ([]threadmodel.Thread, []*threadmodel.Thread, error) {
	parentIDSet := make(map[uint]struct{})
	for _, t := range threads {
		if t.ParentID != nil {
			parentIDSet[*t.ParentID] = struct{}{}
		}
	}
	parentMap := make(map[uint]*threadmodel.Thread)
	if len(parentIDSet) > 0 {
		pids := make([]uint, 0, len(parentIDSet))
		for id := range parentIDSet {
			pids = append(pids, id)
		}
		var parentThreads []threadmodel.Thread
		if err := applyThreadVisibility(preloadThreadParentDetails(r.db.readDB), requesterID).
			Where("threads.thread_id IN ?", pids).
			Find(&parentThreads).Error; err != nil {
			return nil, nil, err
		}
		for i := range parentThreads {
			parentMap[parentThreads[i].ThreadID] = &parentThreads[i]
		}
	}

	outThreads := make([]threadmodel.Thread, 0, len(threads))
	parents := make([]*threadmodel.Thread, 0, len(threads))
	for _, thread := range threads {
		var parent *threadmodel.Thread
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
