package threadrepo

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	communitymodel "github.com/sunibgram/domain/model/community"
	threadmodel "github.com/sunibgram/domain/model/thread"
	threadmediamodel "github.com/sunibgram/domain/model/thread_media"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
)

func (r *threadCached) FindByID(id uint) (*threadmodel.Thread, error) {
	ctx := context.Background()
	var thread threadmodel.Thread
	key := ThreadKey(id)
	if err := r.cache.Get(ctx, key, &thread); err != nil {
		if err := PreloadThreadDetails(r.db.readDB).Where("threads.thread_id = ?", id).First(&thread).Error; err != nil {
			return nil, err
		}
		r.cache.Set(ctx, key, thread, r.ttl)
	}
	return &thread, nil
}

func (r *threadCached) FindRoots(requesterID *uint, page, limit int) ([]threadmodel.Thread, int64, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s:viewer:%s", threadRootsKey(page, limit), ThreadVisibilityCachePart(requesterID))
	rootTypes := forYouFeedTypes(true)

	if entry, ok := r.loadFeedIDs(ctx, key); ok {
		threads, err := r.hydrateThreads(ctx, entry.IDs)
		if err != nil {
			return nil, 0, err
		}
		total, ok := r.getCachedCount(ctx, key)
		if !ok {
			countQuery := ApplyThreadVisibility(r.db.readDB.Model(&threadmodel.Thread{}), requesterID).
				Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
				Where("threads.type IN ?", rootTypes).
				Where("threads.community_id IS NULL OR communities.visibility = ?", communitymodel.CommunityVisibilityPublic)
			countQuery = applyMainFeedReplyInteractionFilter(countQuery)
			if err := countQuery.Count(&total).Error; err != nil {
				return nil, 0, err
			}
			r.setCachedCount(ctx, key, total)
		}
		return threads, total, nil
	}

	var threads []threadmodel.Thread
	var total int64
	countQuery := ApplyThreadVisibility(r.db.readDB.Model(&threadmodel.Thread{}), requesterID).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Where("threads.type IN ?", rootTypes).
		Where("threads.community_id IS NULL OR communities.visibility = ?", communitymodel.CommunityVisibilityPublic)
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
	query := ApplyThreadVisibility(PreloadThreadDetails(r.db.readDB), requesterID).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Where("threads.type IN ?", rootTypes).
		Where("threads.community_id IS NULL OR communities.visibility = ?", communitymodel.CommunityVisibilityPublic)
	query = applyMainFeedReplyInteractionFilter(query)
	if err := query.
		Order("threads.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&threads).Error; err != nil {
		return nil, 0, err
	}
	r.storeFeedIDs(ctx, key, threads, nil, ThreadFeedTTL)
	r.cache.RegisterCacheKey(ctx, infrarepoShared.GlobalThreadFeedsTag(), key)
	r.cache.RegisterCacheKey(ctx, infrarepoShared.ThreadRootsGlobalTag(), key)
	if requesterID != nil && *requesterID != 0 {
		r.cache.RegisterCacheKey(ctx, infrarepoShared.ViewerThreadFeedsTag(*requesterID), key)
	}
	return threads, total, nil
}

func (r *threadCached) FindChain(anchorID uint) (threadmodel.Thread, []threadmodel.Thread, []threadmodel.Thread, error) {
	var anchor threadmodel.Thread
	if err := preloadThreadParentDetails(r.db.readDB).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Where("threads.thread_id = ?", anchorID).
		Where("threads.community_id IS NULL OR communities.visibility = ?", communitymodel.CommunityVisibilityPublic).
		First(&anchor).Error; err != nil {
		return threadmodel.Thread{}, nil, nil, err
	}

	ancestors, err := r.findAncestors(anchor)
	if err != nil {
		return threadmodel.Thread{}, nil, nil, err
	}

	children, err := r.findChainChildren(anchorID, anchor.UserID)
	if err != nil {
		return threadmodel.Thread{}, nil, nil, err
	}
	return anchor, ancestors, children, nil
}

// findChainChildren returns the anchor's direct children PLUS the owner
// self-thread continuation — i.e. descendants reached by following parent_id
// links where every step (parent in the chain AND the child) is authored by
// the anchor's owner. This surfaces X-style self-reply chains in one call.
func (r *threadCached) findChainChildren(anchorID, anchorUserID uint) ([]threadmodel.Thread, error) {
	var childIDs []uint
	const cteSQL = `
		WITH RECURSIVE owner_descendants AS (
			SELECT thread_id, parent_id, user_id, 0 AS depth, created_at
			FROM threads
			WHERE parent_id = ?
			UNION ALL
			SELECT t.thread_id, t.parent_id, t.user_id, od.depth + 1, t.created_at
			FROM threads t
			JOIN owner_descendants od ON t.parent_id = od.thread_id
			WHERE od.user_id = ?
			  AND t.user_id = ?
			  AND od.depth < 100
		)
		SELECT thread_id FROM owner_descendants ORDER BY depth, created_at
	`
	if err := r.db.readDB.Raw(cteSQL, anchorID, anchorUserID, anchorUserID).Scan(&childIDs).Error; err != nil {
		return nil, err
	}
	if len(childIDs) == 0 {
		return nil, nil
	}

	var rows []threadmodel.Thread
	if err := preloadThreadParentDetails(r.db.readDB).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Where("threads.thread_id IN ?", childIDs).
		Where("threads.community_id IS NULL OR communities.visibility = ?", communitymodel.CommunityVisibilityPublic).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	byID := make(map[uint]threadmodel.Thread, len(rows))
	for _, t := range rows {
		byID[t.ThreadID] = t
	}
	ordered := make([]threadmodel.Thread, 0, len(childIDs))
	for _, id := range childIDs {
		if t, ok := byID[id]; ok {
			ordered = append(ordered, t)
		}
	}
	return ordered, nil
}

func (r *threadCached) findAncestors(anchor threadmodel.Thread) ([]threadmodel.Thread, error) {
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

	var rows []threadmodel.Thread
	if err := preloadThreadParentDetails(r.db.readDB).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Where("threads.thread_id IN ?", ancestorIDs).
		Where("threads.community_id IS NULL OR communities.visibility = ?", communitymodel.CommunityVisibilityPublic).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	byID := make(map[uint]threadmodel.Thread, len(rows))
	for _, t := range rows {
		byID[t.ThreadID] = t
	}
	ordered := make([]threadmodel.Thread, 0, len(ancestorIDs))
	for _, id := range ancestorIDs {
		if t, ok := byID[id]; ok {
			ordered = append(ordered, t)
		}
	}
	return ordered, nil
}

func (r *threadCached) FindRepliesContent(parentID uint, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 50
	}
	var out []string
	err := r.db.readDB.Model(&threadmodel.Thread{}).
		Where("parent_id = ?", parentID).
		Order("like_count DESC, created_at DESC").
		Limit(limit).
		Pluck("content", &out).Error
	return out, err
}

func (r *threadCached) FindByRootID(rootID uint, page, limit int) ([]threadmodel.Thread, int64, error) {
	const maxFindByRootIDRows = 50

	if limit > maxFindByRootIDRows {
		limit = maxFindByRootIDRows
	}

	ctx := context.Background()
	key := threadChainKey(rootID, page, limit)

	countLoader := func() (int64, error) {
		var total int64
		if err := r.db.readDB.Model(&threadmodel.Thread{}).
			Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
			Where("threads.thread_id = ? OR threads.root_id = ?", rootID, rootID).
			Where("threads.community_id IS NULL OR communities.visibility = ?", communitymodel.CommunityVisibilityPublic).
			Count(&total).Error; err != nil {
			return 0, err
		}
		return total, nil
	}

	if entry, ok := r.loadFeedIDs(ctx, key); ok {
		// FindByRootID hydrates via preloadThreadParentDetails, lighter than
		// the default PreloadThreadDetails used by hydrateThreads. Load missing
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

	var threads []threadmodel.Thread
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
		Where("threads.community_id IS NULL OR communities.visibility = ?", communitymodel.CommunityVisibilityPublic).
		Order("threads.like_count DESC, threads.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&threads).Error; err != nil {
		return nil, 0, err
	}
	r.storeChainFeedIDs(ctx, key, threads)
	r.cache.RegisterCacheKey(ctx, infrarepoShared.GlobalThreadFeedsTag(), key)
	r.cache.RegisterCacheKey(ctx, infrarepoShared.ThreadChainGlobalTag(), key)
	r.cache.RegisterCacheKey(ctx, infrarepoShared.ThreadChainTag(rootID), key)
	return threads, total, nil
}

func (r *threadCached) FindCommunityTop(communityID, requesterUserID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	clause, _ := communitySortClause("top")
	return r.findCommunityThreads(communityID, requesterUserID, page, limit, "top", clause)
}

func (r *threadCached) FindCommunityLatest(communityID, requesterUserID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	clause, _ := communitySortClause("latest")
	return r.findCommunityThreads(communityID, requesterUserID, page, limit, "latest", clause)
}

func (r *threadCached) FindCommunityThreadsBySort(communityID, requesterUserID uint, sort string, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
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

func (r *threadCached) findCommunityThreads(communityID, requesterUserID uint, page, limit int, keySuffix, orderClause string) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s:viewer:%d", threadCommunityKey(communityID, page, limit, keySuffix), requesterUserID)
	requesterID := &requesterUserID
	threadTypes := []string{string(threadmodel.ThreadTypePost), string(threadmodel.ThreadTypeRepost), string(threadmodel.ThreadTypeReply)}

	countLoader := func() (int64, error) {
		var total int64
		if err := ApplyThreadVisibility(r.db.readDB.Model(&threadmodel.Thread{}), requesterID).
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

	var threads []threadmodel.Thread
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
	if err := ApplyThreadVisibility(PreloadThreadDetails(r.db.readDB), requesterID).
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

	r.storeFeedIDs(ctx, key, threads, parents, ThreadFeedTTL)
	r.cache.RegisterCacheKey(ctx, infrarepoShared.GlobalThreadFeedsTag(), key)
	r.cache.RegisterCacheKey(ctx, infrarepoShared.ThreadCommunityGlobalTag(), key)
	r.cache.RegisterCacheKey(ctx, infrarepoShared.ThreadCommunityTag(communityID), key)
	if requesterUserID != 0 {
		r.cache.RegisterCacheKey(ctx, infrarepoShared.ViewerThreadFeedsTag(requesterUserID), key)
	}
	return threads, parents, total, nil
}

func (r *threadCached) FindCommunityMedia(communityID, requesterUserID uint) ([]threadmediamodel.CommunityThreadMedia, error) {
	var rows []threadmediamodel.ThreadMedia
	requesterID := &requesterUserID
	if err := ApplyThreadVisibility(r.db.readDB.
		Table("thread_media").
		Select("thread_media.media_id, thread_media.thread_id, thread_media.object_name, thread_media.created_at").
		Joins("JOIN threads ON threads.thread_id = thread_media.thread_id"), requesterID).
		Joins("LEFT JOIN threads original ON original.thread_id = threads.original_thread_id").
		Where("COALESCE(threads.community_id, original.community_id) = ?", communityID).
		Where("threads.type IN ?", []string{string(threadmodel.ThreadTypePost), string(threadmodel.ThreadTypeRepost), string(threadmodel.ThreadTypeReply)}).
		Order("threads.created_at DESC, thread_media.created_at ASC").
		Limit(500).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	byThread := make(map[uint]int)
	result := make([]threadmediamodel.CommunityThreadMedia, 0)
	for _, media := range rows {
		idx, ok := byThread[media.ThreadID]
		if !ok {
			idx = len(result)
			byThread[media.ThreadID] = idx
			result = append(result, threadmediamodel.CommunityThreadMedia{ThreadID: media.ThreadID})
		}
		result[idx].Media = append(result[idx].Media, media)
	}
	return result, nil
}

func (r *threadCached) FindMyCommunityThreads(userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	ctx := context.Background()
	key := fmt.Sprintf("thread:my-communities:user:%d:page:%d:limit:%d", userID, page, limit)
	requesterID := &userID
	threadTypes := []string{string(threadmodel.ThreadTypePost), string(threadmodel.ThreadTypeRepost), string(threadmodel.ThreadTypeReply)}

	countLoader := func() (int64, error) {
		var total int64
		base := ApplyThreadVisibility(r.db.readDB.Model(&threadmodel.Thread{}), requesterID).
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

	var threads []threadmodel.Thread
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
	if err := ApplyThreadVisibility(PreloadThreadDetails(r.db.readDB), requesterID).
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

	r.storeFeedIDs(ctx, key, threads, parents, ThreadFeedTTL)
	r.cache.RegisterCacheKey(ctx, infrarepoShared.GlobalThreadFeedsTag(), key)
	r.cache.RegisterCacheKey(ctx, infrarepoShared.ThreadMyCommunitiesGlobalTag(), key)
	if userID != 0 {
		r.cache.RegisterCacheKey(ctx, infrarepoShared.ViewerThreadFeedsTag(userID), key)
	}
	return threads, parents, total, nil
}

func (r *threadCached) FindAllCommunityThreads(userID uint, categoryID *uuid.UUID, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	ctx := context.Background()
	catStr := "all"
	if categoryID != nil {
		catStr = categoryID.String()
	}
	key := fmt.Sprintf("thread:all-communities:user:%d:cat:%s:page:%d:limit:%d", userID, catStr, page, limit)

	countLoader := func() (int64, error) {
		var total int64
		base := r.db.readDB.Model(&threadmodel.Thread{}).
			Joins("LEFT JOIN threads original ON original.thread_id = threads.original_thread_id").
			Joins("JOIN communities ON communities.community_id = COALESCE(threads.community_id, original.community_id)").
			Joins("LEFT JOIN user_settings us ON us.user_id = threads.user_id").
			Joins("LEFT JOIN user_relationships ur ON ur.follower_id = ? AND ur.following_id = threads.user_id", userID).
			Where("COALESCE(threads.community_id, original.community_id) IS NOT NULL").
			Where("threads.type IN ?", []string{string(threadmodel.ThreadTypePost), string(threadmodel.ThreadTypeRepost), string(threadmodel.ThreadTypeReply)}).
			Where("communities.is_active = ?", true).
			Where("communities.visibility = ? OR communities.community_id IN (SELECT community_id FROM community_members WHERE user_id = ?)",
				communitymodel.CommunityVisibilityPublic, userID).
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

	var threads []threadmodel.Thread
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
	query := PreloadThreadDetails(r.db.readDB).
		Joins("LEFT JOIN threads original ON original.thread_id = threads.original_thread_id").
		Joins("JOIN communities ON communities.community_id = COALESCE(threads.community_id, original.community_id)").
		Joins("LEFT JOIN user_settings us ON us.user_id = threads.user_id").
		Joins("LEFT JOIN user_relationships ur ON ur.follower_id = ? AND ur.following_id = threads.user_id", userID).
		Where("COALESCE(threads.community_id, original.community_id) IS NOT NULL").
		Where("threads.type IN ?", []string{string(threadmodel.ThreadTypePost), string(threadmodel.ThreadTypeRepost), string(threadmodel.ThreadTypeReply)}).
		Where("communities.is_active = ?", true).
		Where("communities.visibility = ? OR communities.community_id IN (SELECT community_id FROM community_members WHERE user_id = ?)",
			communitymodel.CommunityVisibilityPublic, userID).
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

	r.storeFeedIDs(ctx, key, threads, parents, ThreadFeedTTL)
	r.cache.RegisterCacheKey(ctx, infrarepoShared.GlobalThreadFeedsTag(), key)
	r.cache.RegisterCacheKey(ctx, infrarepoShared.ThreadAllCommunitiesGlobalTag(), key)
	if userID != 0 {
		r.cache.RegisterCacheKey(ctx, infrarepoShared.ViewerThreadFeedsTag(userID), key)
	}
	return threads, parents, total, nil
}

func (r *threadCached) FindAllCommunityThreadsByQueryTop(userID uint, query string, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	return r.findAllCommunityThreadsByQuery(
		userID,
		query,
		page,
		limit,
		"top",
		"(threads.like_count + threads.repost_count + threads.bookmark_count) DESC, threads.created_at DESC",
	)
}

func (r *threadCached) FindAllCommunityThreadsByQueryLatest(userID uint, query string, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	return r.findAllCommunityThreadsByQuery(userID, query, page, limit, "latest", "threads.created_at DESC")
}

func (r *threadCached) findAllCommunityThreadsByQuery(userID uint, query string, page, limit int, keySuffix, orderClause string) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	ctx := context.Background()
	query = strings.TrimSpace(query)
	queryLike := "%" + query + "%"
	key := communityThreadsQueryKey(keySuffix, strings.ToLower(query), userID, page, limit)

	countLoader := func() (int64, error) {
		var total int64
		base := r.db.readDB.Model(&threadmodel.Thread{}).
			Joins("LEFT JOIN threads original ON original.thread_id = threads.original_thread_id").
			Joins("JOIN communities ON communities.community_id = COALESCE(threads.community_id, original.community_id)").
			Joins("LEFT JOIN user_settings us ON us.user_id = threads.user_id").
			Joins("LEFT JOIN user_relationships ur ON ur.follower_id = ? AND ur.following_id = threads.user_id", userID).
			Where("COALESCE(threads.community_id, original.community_id) IS NOT NULL").
			Where("threads.type IN ?", []string{string(threadmodel.ThreadTypePost), string(threadmodel.ThreadTypeRepost), string(threadmodel.ThreadTypeReply)}).
			Where("communities.is_active = ?", true).
			Where("(communities.name ILIKE ? OR COALESCE(NULLIF(threads.content, ''), original.content) ILIKE ?)", queryLike, queryLike).
			Where("communities.visibility = ? OR communities.community_id IN (SELECT community_id FROM community_members WHERE user_id = ?)",
				communitymodel.CommunityVisibilityPublic, userID).
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

	var threads []threadmodel.Thread
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
	if err := PreloadThreadDetails(r.db.readDB).
		Joins("LEFT JOIN threads original ON original.thread_id = threads.original_thread_id").
		Joins("JOIN communities ON communities.community_id = COALESCE(threads.community_id, original.community_id)").
		Joins("LEFT JOIN user_settings us ON us.user_id = threads.user_id").
		Joins("LEFT JOIN user_relationships ur ON ur.follower_id = ? AND ur.following_id = threads.user_id", userID).
		Where("COALESCE(threads.community_id, original.community_id) IS NOT NULL").
		Where("threads.type IN ?", []string{string(threadmodel.ThreadTypePost), string(threadmodel.ThreadTypeRepost), string(threadmodel.ThreadTypeReply)}).
		Where("communities.is_active = ?", true).
		Where("(communities.name ILIKE ? OR COALESCE(NULLIF(threads.content, ''), original.content) ILIKE ?)", queryLike, queryLike).
		Where("communities.visibility = ? OR communities.community_id IN (SELECT community_id FROM community_members WHERE user_id = ?)",
			communitymodel.CommunityVisibilityPublic, userID).
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

	r.storeFeedIDs(ctx, key, threads, parents, ThreadFeedTTL)
	r.cache.RegisterCacheKey(ctx, infrarepoShared.GlobalThreadFeedsTag(), key)
	r.cache.RegisterCacheKey(ctx, infrarepoShared.CommunityThreadsGlobalTag(), key)
	if userID != 0 {
		r.cache.RegisterCacheKey(ctx, infrarepoShared.ViewerThreadFeedsTag(userID), key)
	}
	return threads, parents, total, nil
}
