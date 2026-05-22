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
		if err := preloadThreadDetails(r.db.readDB).Where("threads.thread_id = ?", id).First(&thread).Error; err != nil {
			return nil, err
		}
		r.cache.Set(ctx, key, thread, r.ttl)
	}
	return &thread, nil
}

func (r *threadCached) FindRoots(requesterID *uint, page, limit int) ([]model.Thread, int64, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s:viewer:%s", threadRootsKey(page, limit), threadVisibilityCachePart(requesterID))
	rootTypes := forYouFeedTypes(true)

	if entry, ok := r.loadFeedIDs(ctx, key); ok {
		threads, err := r.hydrateThreads(ctx, entry.IDs)
		if err != nil {
			return nil, 0, err
		}
		total, ok := r.getCachedCount(ctx, key)
		if !ok {
			countQuery := applyThreadVisibility(r.db.readDB.Model(&model.Thread{}), requesterID).
				Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
				Where("threads.type IN ?", rootTypes).
				Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic)
			countQuery = applyMainFeedReplyInteractionFilter(countQuery)
			if err := countQuery.Count(&total).Error; err != nil {
				return nil, 0, err
			}
			r.setCachedCount(ctx, key, total)
		}
		return threads, total, nil
	}

	var threads []model.Thread
	var total int64
	countQuery := applyThreadVisibility(r.db.readDB.Model(&model.Thread{}), requesterID).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Where("threads.type IN ?", rootTypes).
		Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic)
	countQuery = applyMainFeedReplyInteractionFilter(countQuery)
	if cachedTotal, ok := r.getCachedCount(ctx, key); ok {
		total = cachedTotal
	} else {
		if err := countQuery.Count(&total).Error; err != nil {
			return nil, 0, err
		}
		r.setCachedCount(ctx, key, total)
	}
	offset := (page - 1) * limit
	query := applyThreadVisibility(preloadThreadDetails(r.db.readDB), requesterID).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Where("threads.type IN ?", rootTypes).
		Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic)
	query = applyMainFeedReplyInteractionFilter(query)
	if err := query.
		Order("threads.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&threads).Error; err != nil {
		return nil, 0, err
	}
	r.storeFeedIDs(ctx, key, threads, nil, threadFeedTTL)
	r.cache.RegisterCacheKey(ctx, globalThreadFeedsTag(), key)
	r.cache.RegisterCacheKey(ctx, threadRootsGlobalTag(), key)
	if requesterID != nil && *requesterID != 0 {
		r.cache.RegisterCacheKey(ctx, viewerThreadFeedsTag(*requesterID), key)
	}
	return threads, total, nil
}

func (r *threadCached) FindChain(anchorID uint) (model.Thread, []model.Thread, []model.Thread, error) {
	var anchor model.Thread
	if err := preloadThreadParentDetails(r.db.readDB).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Where("threads.thread_id = ?", anchorID).
		Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
		First(&anchor).Error; err != nil {
		return model.Thread{}, nil, nil, err
	}

	ancestors, err := r.findAncestors(anchor)
	if err != nil {
		return model.Thread{}, nil, nil, err
	}

	var children []model.Thread
	if err := preloadThreadParentDetails(r.db.readDB).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Where("threads.parent_id = ?", anchorID).
		Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
		Order("threads.created_at ASC").
		Find(&children).Error; err != nil {
		return model.Thread{}, nil, nil, err
	}
	return anchor, ancestors, children, nil
}

func (r *threadCached) findAncestors(anchor model.Thread) ([]model.Thread, error) {
	if anchor.ParentID == nil {
		return nil, nil
	}

	var ancestorIDs []uint
	cteSQL := `
		WITH RECURSIVE ancestors AS (
			SELECT thread_id, parent_id, 0 AS depth
			FROM threads
			WHERE thread_id = ?
			UNION ALL
			SELECT t.thread_id, t.parent_id, a.depth + 1
			FROM threads t
			JOIN ancestors a ON t.thread_id = a.parent_id
			WHERE a.depth < 100
		)
		SELECT thread_id FROM ancestors ORDER BY depth DESC
	`
	if err := r.db.readDB.Raw(cteSQL, *anchor.ParentID).Scan(&ancestorIDs).Error; err != nil {
		return nil, err
	}
	if len(ancestorIDs) == 0 {
		return nil, nil
	}

	var rows []model.Thread
	if err := preloadThreadParentDetails(r.db.readDB).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Where("threads.thread_id IN ?", ancestorIDs).
		Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	byID := make(map[uint]model.Thread, len(rows))
	for _, t := range rows {
		byID[t.ThreadID] = t
	}
	ordered := make([]model.Thread, 0, len(ancestorIDs))
	for _, id := range ancestorIDs {
		if t, ok := byID[id]; ok {
			ordered = append(ordered, t)
		}
	}
	return ordered, nil
}

func (r *threadCached) FindByRootID(rootID uint, page, limit int) ([]model.Thread, int64, error) {
	const maxFindByRootIDRows = 50

	if limit > maxFindByRootIDRows {
		limit = maxFindByRootIDRows
	}

	ctx := context.Background()
	key := threadChainKey(rootID, page, limit)

	countLoader := func() (int64, error) {
		var total int64
		if err := r.db.readDB.Model(&model.Thread{}).
			Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
			Where("threads.thread_id = ? OR threads.root_id = ?", rootID, rootID).
			Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
			Count(&total).Error; err != nil {
			return 0, err
		}
		return total, nil
	}

	if entry, ok := r.loadFeedIDs(ctx, key); ok {
		// FindByRootID hydrates via preloadThreadParentDetails, lighter than
		// the default preloadThreadDetails used by hydrateThreads. Load missing
		// rows ourselves through the parent-detail preload and fall back to
		// per-thread cache otherwise.
		threads, err := r.hydrateChainThreads(ctx, entry.IDs)
		if err != nil {
			return nil, 0, err
		}
		total, ok := r.getCachedCount(ctx, key)
		if !ok {
			total, err = countLoader()
			if err != nil {
				return nil, 0, err
			}
			r.setCachedCount(ctx, key, total)
		}
		return threads, total, nil
	}

	var threads []model.Thread
	total, ok := r.getCachedCount(ctx, key)
	if !ok {
		var err error
		total, err = countLoader()
		if err != nil {
			return nil, 0, err
		}
		r.setCachedCount(ctx, key, total)
	}
	offset := (page - 1) * limit
	if err := preloadThreadParentDetails(r.db.readDB).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Where("threads.thread_id = ? OR threads.root_id = ?", rootID, rootID).
		Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
		Order("threads.like_count DESC, threads.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&threads).Error; err != nil {
		return nil, 0, err
	}
	r.storeChainFeedIDs(ctx, key, threads)
	r.cache.RegisterCacheKey(ctx, globalThreadFeedsTag(), key)
	r.cache.RegisterCacheKey(ctx, threadChainGlobalTag(), key)
	r.cache.RegisterCacheKey(ctx, threadChainTag(rootID), key)
	return threads, total, nil
}

func (r *threadCached) FindCommunityTop(communityID, requesterUserID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	clause, _ := communitySortClause("top")
	return r.findCommunityThreads(communityID, requesterUserID, page, limit, "top", clause)
}

func (r *threadCached) FindCommunityLatest(communityID, requesterUserID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	clause, _ := communitySortClause("latest")
	return r.findCommunityThreads(communityID, requesterUserID, page, limit, "latest", clause)
}

func (r *threadCached) FindCommunityThreadsBySort(communityID, requesterUserID uint, sort string, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	orderClause, keySuffix := communitySortClause(sort)
	return r.findCommunityThreads(communityID, requesterUserID, page, limit, keySuffix, orderClause)
}

func communitySortClause(sort string) (string, string) {
	switch sort {
	case "top":
		return "(COALESCE(original.like_count, threads.like_count) + COALESCE(original.repost_count, threads.repost_count) + COALESCE(original.bookmark_count, threads.bookmark_count)) DESC, threads.created_at DESC", "top"
	case "likes":
		return "COALESCE(original.like_count, threads.like_count) DESC, threads.created_at DESC", "likes"
	default:
		return "threads.created_at DESC", "latest"
	}
}

func (r *threadCached) findCommunityThreads(communityID, requesterUserID uint, page, limit int, keySuffix, orderClause string) ([]model.Thread, []*model.Thread, int64, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s:viewer:%d", threadCommunityKey(communityID, page, limit, keySuffix), requesterUserID)
	requesterID := &requesterUserID
	threadTypes := []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost), string(model.ThreadTypeReply)}

	countLoader := func() (int64, error) {
		var total int64
		if err := applyThreadVisibility(r.db.readDB.Model(&model.Thread{}), requesterID).
			Joins("LEFT JOIN threads original ON original.thread_id = threads.original_thread_id").
			Where("COALESCE(threads.community_id, original.community_id) = ?", communityID).
			Where("threads.type IN ?", threadTypes).
			Count(&total).Error; err != nil {
			return 0, err
		}
		return total, nil
	}

	if entry, ok := r.loadFeedIDs(ctx, key); ok && len(entry.ParentIDs) == len(entry.IDs) {
		threads, err := r.hydrateThreads(ctx, entry.IDs)
		if err != nil {
			return nil, nil, 0, err
		}
		parents, err := r.hydrateParents(ctx, entry.ParentIDs)
		if err != nil {
			return nil, nil, 0, err
		}
		total, ok := r.getCachedCount(ctx, key)
		if !ok {
			total, err = countLoader()
			if err != nil {
				return nil, nil, 0, err
			}
			r.setCachedCount(ctx, key, total)
		}
		return threads, parents, total, nil
	}

	var threads []model.Thread
	total, ok := r.getCachedCount(ctx, key)
	if !ok {
		var err error
		total, err = countLoader()
		if err != nil {
			return nil, nil, 0, err
		}
		r.setCachedCount(ctx, key, total)
	}
	offset := (page - 1) * limit
	if err := applyThreadVisibility(preloadThreadDetails(r.db.readDB), requesterID).
		Joins("LEFT JOIN threads original ON original.thread_id = threads.original_thread_id").
		Where("COALESCE(threads.community_id, original.community_id) = ?", communityID).
		Where("threads.type IN ?", threadTypes).
		Order(orderClause).
		Offset(offset).Limit(limit).
		Find(&threads).Error; err != nil {
		return nil, nil, 0, err
	}

	threads, parents, err := r.attachVisibleParents(threads, requesterID)
	if err != nil {
		return nil, nil, 0, err
	}

	r.storeFeedIDs(ctx, key, threads, parents, threadFeedTTL)
	r.cache.RegisterCacheKey(ctx, globalThreadFeedsTag(), key)
	r.cache.RegisterCacheKey(ctx, threadCommunityGlobalTag(), key)
	r.cache.RegisterCacheKey(ctx, threadCommunityTag(communityID), key)
	if requesterUserID != 0 {
		r.cache.RegisterCacheKey(ctx, viewerThreadFeedsTag(requesterUserID), key)
	}
	return threads, parents, total, nil
}

func (r *threadCached) FindCommunityMedia(communityID, requesterUserID uint) ([]model.CommunityThreadMedia, error) {
	var rows []model.ThreadMedia
	requesterID := &requesterUserID
	if err := applyThreadVisibility(r.db.readDB.
		Table("thread_media").
		Select("thread_media.media_id, thread_media.thread_id, thread_media.object_name, thread_media.bucket, thread_media.created_at").
		Joins("JOIN threads ON threads.thread_id = thread_media.thread_id"), requesterID).
		Joins("LEFT JOIN threads original ON original.thread_id = threads.original_thread_id").
		Where("COALESCE(threads.community_id, original.community_id) = ?", communityID).
		Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost), string(model.ThreadTypeReply)}).
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

func (r *threadCached) FindMyCommunityThreads(userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	ctx := context.Background()
	key := fmt.Sprintf("thread:my-communities:user:%d:page:%d:limit:%d", userID, page, limit)
	requesterID := &userID
	threadTypes := []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost), string(model.ThreadTypeReply)}

	countLoader := func() (int64, error) {
		var total int64
		base := applyThreadVisibility(r.db.readDB.Model(&model.Thread{}), requesterID).
			Joins("LEFT JOIN threads original ON original.thread_id = threads.original_thread_id").
			Joins("JOIN communities ON communities.community_id = COALESCE(threads.community_id, original.community_id)").
			Joins("JOIN community_members cm ON cm.community_id = COALESCE(threads.community_id, original.community_id) AND cm.user_id = ?", userID).
			Where("COALESCE(threads.community_id, original.community_id) IS NOT NULL").
			Where("threads.type IN ?", threadTypes).
			Where("communities.is_active = ?", true)
		if err := base.Count(&total).Error; err != nil {
			return 0, err
		}
		return total, nil
	}

	if entry, ok := r.loadFeedIDs(ctx, key); ok && len(entry.ParentIDs) == len(entry.IDs) {
		threads, err := r.hydrateThreads(ctx, entry.IDs)
		if err != nil {
			return nil, nil, 0, err
		}
		parents, err := r.hydrateParents(ctx, entry.ParentIDs)
		if err != nil {
			return nil, nil, 0, err
		}
		total, ok := r.getCachedCount(ctx, key)
		if !ok {
			total, err = countLoader()
			if err != nil {
				return nil, nil, 0, err
			}
			r.setCachedCount(ctx, key, total)
		}
		return threads, parents, total, nil
	}

	var threads []model.Thread
	total, ok := r.getCachedCount(ctx, key)
	if !ok {
		var err error
		total, err = countLoader()
		if err != nil {
			return nil, nil, 0, err
		}
		r.setCachedCount(ctx, key, total)
	}

	offset := (page - 1) * limit
	if err := applyThreadVisibility(preloadThreadDetails(r.db.readDB), requesterID).
		Joins("LEFT JOIN threads original ON original.thread_id = threads.original_thread_id").
		Joins("JOIN communities ON communities.community_id = COALESCE(threads.community_id, original.community_id)").
		Joins("JOIN community_members cm ON cm.community_id = COALESCE(threads.community_id, original.community_id) AND cm.user_id = ?", userID).
		Where("COALESCE(threads.community_id, original.community_id) IS NOT NULL").
		Where("threads.type IN ?", threadTypes).
		Where("communities.is_active = ?", true).
		Order("threads.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&threads).Error; err != nil {
		return nil, nil, 0, err
	}

	threads, parents, err := r.attachVisibleParents(threads, requesterID)
	if err != nil {
		return nil, nil, 0, err
	}

	r.storeFeedIDs(ctx, key, threads, parents, threadFeedTTL)
	r.cache.RegisterCacheKey(ctx, globalThreadFeedsTag(), key)
	r.cache.RegisterCacheKey(ctx, threadMyCommunitiesGlobalTag(), key)
	if userID != 0 {
		r.cache.RegisterCacheKey(ctx, viewerThreadFeedsTag(userID), key)
	}
	return threads, parents, total, nil
}

func (r *threadCached) FindAllCommunityThreads(userID uint, categoryID *uuid.UUID, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	ctx := context.Background()
	catStr := "all"
	if categoryID != nil {
		catStr = categoryID.String()
	}
	key := fmt.Sprintf("thread:all-communities:user:%d:cat:%s:page:%d:limit:%d", userID, catStr, page, limit)

	countLoader := func() (int64, error) {
		var total int64
		base := r.db.readDB.Model(&model.Thread{}).
			Joins("LEFT JOIN threads original ON original.thread_id = threads.original_thread_id").
			Joins("JOIN communities ON communities.community_id = COALESCE(threads.community_id, original.community_id)").
			Joins("LEFT JOIN user_settings us ON us.user_id = threads.user_id").
			Joins("LEFT JOIN user_relationships ur ON ur.follower_id = ? AND ur.following_id = threads.user_id", userID).
			Where("COALESCE(threads.community_id, original.community_id) IS NOT NULL").
			Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost), string(model.ThreadTypeReply)}).
			Where("communities.is_active = ?", true).
			Where("communities.visibility = ? OR communities.community_id IN (SELECT community_id FROM community_members WHERE user_id = ?)",
				model.CommunityVisibilityPublic, userID).
			Where("threads.user_id = ? OR COALESCE(us.is_private, false) = false OR ur.following_id IS NOT NULL", userID)
		if categoryID != nil {
			base = base.Where("communities.category_id = ?", categoryID)
		}
		if err := base.Count(&total).Error; err != nil {
			return 0, err
		}
		return total, nil
	}

	if entry, ok := r.loadFeedIDs(ctx, key); ok && len(entry.ParentIDs) == len(entry.IDs) {
		threads, err := r.hydrateThreads(ctx, entry.IDs)
		if err != nil {
			return nil, nil, 0, err
		}
		parents, err := r.hydrateParents(ctx, entry.ParentIDs)
		if err != nil {
			return nil, nil, 0, err
		}
		total, ok := r.getCachedCount(ctx, key)
		if !ok {
			total, err = countLoader()
			if err != nil {
				return nil, nil, 0, err
			}
			r.setCachedCount(ctx, key, total)
		}
		return threads, parents, total, nil
	}

	var threads []model.Thread
	total, ok := r.getCachedCount(ctx, key)
	if !ok {
		var err error
		total, err = countLoader()
		if err != nil {
			return nil, nil, 0, err
		}
		r.setCachedCount(ctx, key, total)
	}

	offset := (page - 1) * limit
	query := preloadThreadDetails(r.db.readDB).
		Joins("LEFT JOIN threads original ON original.thread_id = threads.original_thread_id").
		Joins("JOIN communities ON communities.community_id = COALESCE(threads.community_id, original.community_id)").
		Joins("LEFT JOIN user_settings us ON us.user_id = threads.user_id").
		Joins("LEFT JOIN user_relationships ur ON ur.follower_id = ? AND ur.following_id = threads.user_id", userID).
		Where("COALESCE(threads.community_id, original.community_id) IS NOT NULL").
		Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost), string(model.ThreadTypeReply)}).
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
		return nil, nil, 0, err
	}

	threads, parents, err := r.attachVisibleParents(threads, &userID)
	if err != nil {
		return nil, nil, 0, err
	}

	r.storeFeedIDs(ctx, key, threads, parents, threadFeedTTL)
	r.cache.RegisterCacheKey(ctx, globalThreadFeedsTag(), key)
	r.cache.RegisterCacheKey(ctx, threadAllCommunitiesGlobalTag(), key)
	if userID != 0 {
		r.cache.RegisterCacheKey(ctx, viewerThreadFeedsTag(userID), key)
	}
	return threads, parents, total, nil
}

func (r *threadCached) FindAllCommunityThreadsByQueryTop(userID uint, query string, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	return r.findAllCommunityThreadsByQuery(
		userID,
		query,
		page,
		limit,
		"top",
		"(threads.like_count + threads.repost_count + threads.bookmark_count) DESC, threads.created_at DESC",
	)
}

func (r *threadCached) FindAllCommunityThreadsByQueryLatest(userID uint, query string, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	return r.findAllCommunityThreadsByQuery(userID, query, page, limit, "latest", "threads.created_at DESC")
}

func (r *threadCached) findAllCommunityThreadsByQuery(userID uint, query string, page, limit int, keySuffix, orderClause string) ([]model.Thread, []*model.Thread, int64, error) {
	ctx := context.Background()
	query = strings.TrimSpace(query)
	queryLike := "%" + query + "%"
	key := communityThreadsQueryKey(keySuffix, strings.ToLower(query), userID, page, limit)

	countLoader := func() (int64, error) {
		var total int64
		base := r.db.readDB.Model(&model.Thread{}).
			Joins("LEFT JOIN threads original ON original.thread_id = threads.original_thread_id").
			Joins("JOIN communities ON communities.community_id = COALESCE(threads.community_id, original.community_id)").
			Joins("LEFT JOIN user_settings us ON us.user_id = threads.user_id").
			Joins("LEFT JOIN user_relationships ur ON ur.follower_id = ? AND ur.following_id = threads.user_id", userID).
			Where("COALESCE(threads.community_id, original.community_id) IS NOT NULL").
			Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost), string(model.ThreadTypeReply)}).
			Where("communities.is_active = ?", true).
			Where("(communities.name ILIKE ? OR COALESCE(NULLIF(threads.content, ''), original.content) ILIKE ?)", queryLike, queryLike).
			Where("communities.visibility = ? OR communities.community_id IN (SELECT community_id FROM community_members WHERE user_id = ?)",
				model.CommunityVisibilityPublic, userID).
			Where("threads.user_id = ? OR COALESCE(us.is_private, false) = false OR ur.following_id IS NOT NULL", userID)
		if err := base.Count(&total).Error; err != nil {
			return 0, err
		}
		return total, nil
	}

	if entry, ok := r.loadFeedIDs(ctx, key); ok && len(entry.ParentIDs) == len(entry.IDs) {
		threads, err := r.hydrateThreads(ctx, entry.IDs)
		if err != nil {
			return nil, nil, 0, err
		}
		parents, err := r.hydrateParents(ctx, entry.ParentIDs)
		if err != nil {
			return nil, nil, 0, err
		}
		total, ok := r.getCachedCount(ctx, key)
		if !ok {
			total, err = countLoader()
			if err != nil {
				return nil, nil, 0, err
			}
			r.setCachedCount(ctx, key, total)
		}
		return threads, parents, total, nil
	}

	var threads []model.Thread
	total, ok := r.getCachedCount(ctx, key)
	if !ok {
		var err error
		total, err = countLoader()
		if err != nil {
			return nil, nil, 0, err
		}
		r.setCachedCount(ctx, key, total)
	}

	offset := (page - 1) * limit
	if err := preloadThreadDetails(r.db.readDB).
		Joins("LEFT JOIN threads original ON original.thread_id = threads.original_thread_id").
		Joins("JOIN communities ON communities.community_id = COALESCE(threads.community_id, original.community_id)").
		Joins("LEFT JOIN user_settings us ON us.user_id = threads.user_id").
		Joins("LEFT JOIN user_relationships ur ON ur.follower_id = ? AND ur.following_id = threads.user_id", userID).
		Where("COALESCE(threads.community_id, original.community_id) IS NOT NULL").
		Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost), string(model.ThreadTypeReply)}).
		Where("communities.is_active = ?", true).
		Where("(communities.name ILIKE ? OR COALESCE(NULLIF(threads.content, ''), original.content) ILIKE ?)", queryLike, queryLike).
		Where("communities.visibility = ? OR communities.community_id IN (SELECT community_id FROM community_members WHERE user_id = ?)",
			model.CommunityVisibilityPublic, userID).
		Where("threads.user_id = ? OR COALESCE(us.is_private, false) = false OR ur.following_id IS NOT NULL", userID).
		Order(orderClause).
		Offset(offset).Limit(limit).
		Find(&threads).Error; err != nil {
		return nil, nil, 0, err
	}

	threads, parents, err := r.attachVisibleParents(threads, &userID)
	if err != nil {
		return nil, nil, 0, err
	}

	r.storeFeedIDs(ctx, key, threads, parents, threadFeedTTL)
	r.cache.RegisterCacheKey(ctx, globalThreadFeedsTag(), key)
	r.cache.RegisterCacheKey(ctx, communityThreadsGlobalTag(), key)
	if userID != 0 {
		r.cache.RegisterCacheKey(ctx, viewerThreadFeedsTag(userID), key)
	}
	return threads, parents, total, nil
}
