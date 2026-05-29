package eventrepo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	communitymodel "github.com/sunibgram/domain/model/community"
	threadmodel "github.com/sunibgram/domain/model/thread"
	eventrepoiface "github.com/sunibgram/domain/repository/event"
	"github.com/sunibgram/infrastructure/cache"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
	threadrepo "github.com/sunibgram/infrastructure/repository/thread"
)

type eventRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
	ttl     time.Duration
}

type cachedEvents struct {
	Trending []threadmodel.Event
	Running  []threadmodel.Event
	Future   []threadmodel.Event
	Past     []threadmodel.Event
}

type cachedEventThreads struct {
	Threads []threadmodel.Thread
	Total   int64
}

func NewEventRepository(writeDB, readDB *gorm.DB, c *cache.Cache) eventrepoiface.IEventRepository {
	return &eventRepository{writeDB: writeDB, readDB: readDB, cache: c, ttl: cache.DefaultTTL}
}

func eventListKey(limit int) string {
	return fmt.Sprintf("event:list:limit:%d", limit)
}

func eventApprovedListKey(limit int) string {
	return fmt.Sprintf("event:approved:list:limit:%d", limit)
}

func eventPendingListKey(limit int) string {
	return fmt.Sprintf("event:pending:list:limit:%d", limit)
}

func eventTopRunningFutureKey(limit int) string {
	return fmt.Sprintf("event:top:running-future:limit:%d", limit)
}

func eventThreadsKey(id, filter, sort, viewer string, page, limit int) string {
	return fmt.Sprintf("event:threads:%s:filter:%s:sort:%s:viewer:%s:page:%d:limit:%d", id, filter, sort, viewer, page, limit)
}

func eventThreadSortClause(sort string) string {
	switch sort {
	case "like":
		return "COALESCE(original.like_count, threads.like_count) DESC, threads.created_at DESC"
	case "top":
		return "(COALESCE(original.like_count, threads.like_count) + COALESCE(original.repost_count, threads.repost_count) + COALESCE(original.bookmark_count, threads.bookmark_count)) DESC, threads.created_at DESC"
	default:
		return "threads.created_at DESC"
	}
}

func (r *eventRepository) Save(event *threadmodel.Event) error {
	if err := r.writeDB.Create(event).Error; err != nil {
		return err
	}
	r.invalidateLists()
	return nil
}

func (r *eventRepository) FindByID(id string) (*threadmodel.Event, error) {
	ctx := context.Background()
	key := infrarepoShared.EventIDKey(id)
	var event threadmodel.Event
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &event); err == nil {
			return &event, nil
		}
	}
	if err := r.readDB.First(&event, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if r.cache != nil {
		_ = r.cache.Set(ctx, key, event, r.ttl)
	}
	return &event, nil
}

func (r *eventRepository) FindGrouped(limit int) ([]threadmodel.Event, []threadmodel.Event, []threadmodel.Event, error) {
	ctx := context.Background()
	key := eventListKey(limit)
	var result cachedEvents
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return result.Trending, result.Running, result.Future, nil
		}
	}

	var trending []threadmodel.Event
	if err := r.readDB.
		Order("(week_thread_count * 3 + week_like_count + week_repost_count * 2) DESC, start_time ASC").
		Limit(limit).
		Find(&trending).Error; err != nil {
		return nil, nil, nil, err
	}

	now := time.Now().UTC()
	var running []threadmodel.Event
	if err := r.readDB.
		Where("start_time <= ? AND end_time >= ?", now, now).
		Order("start_time ASC").
		Limit(limit).
		Find(&running).Error; err != nil {
		return nil, nil, nil, err
	}

	var future []threadmodel.Event
	if err := r.readDB.
		Where("start_time > ?", now).
		Order("start_time ASC").
		Limit(limit).
		Find(&future).Error; err != nil {
		return nil, nil, nil, err
	}

	result = cachedEvents{Trending: trending, Running: running, Future: future}
	if r.cache != nil {
		if err := r.cache.Set(ctx, key, result, r.ttl); err == nil {
			r.cache.RegisterCacheKey(ctx, infrarepoShared.EventListsTag(), key)
		}
	}
	return result.Trending, result.Running, result.Future, nil
}

func (r *eventRepository) FindGroupedApproved(limit int) ([]threadmodel.Event, []threadmodel.Event, []threadmodel.Event, []threadmodel.Event, error) {
	ctx := context.Background()
	key := eventApprovedListKey(limit)
	var result cachedEvents
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return result.Trending, result.Running, result.Future, result.Past, nil
		}
	}

	var trending []threadmodel.Event
	if err := r.readDB.
		Where("status = ?", "approved").
		Order("(week_thread_count * 3 + week_like_count + week_repost_count * 2) DESC, start_time ASC").
		Limit(limit).
		Find(&trending).Error; err != nil {
		return nil, nil, nil, nil, err
	}

	now := time.Now().UTC()
	var running []threadmodel.Event
	if err := r.readDB.
		Where("status = ?", "approved").
		Where("start_time <= ? AND end_time >= ?", now, now).
		Order("start_time ASC").
		Limit(limit).
		Find(&running).Error; err != nil {
		return nil, nil, nil, nil, err
	}

	var future []threadmodel.Event
	if err := r.readDB.
		Where("status = ?", "approved").
		Where("start_time > ?", now).
		Order("start_time ASC").
		Limit(limit).
		Find(&future).Error; err != nil {
		return nil, nil, nil, nil, err
	}

	var past []threadmodel.Event
	if err := r.readDB.
		Where("status = ?", "approved").
		Where("end_time < ?", now).
		Order("end_time DESC").
		Limit(limit).
		Find(&past).Error; err != nil {
		return nil, nil, nil, nil, err
	}

	result = cachedEvents{Trending: trending, Running: running, Future: future, Past: past}
	if r.cache != nil {
		if err := r.cache.Set(ctx, key, result, r.ttl); err == nil {
			r.cache.RegisterCacheKey(ctx, infrarepoShared.EventListsTag(), key)
		}
	}
	return result.Trending, result.Running, result.Future, result.Past, nil
}

func (r *eventRepository) FindTopRunningOrFuture(limit int) ([]threadmodel.Event, error) {
	ctx := context.Background()
	key := eventTopRunningFutureKey(limit)
	var events []threadmodel.Event
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &events); err == nil {
			return events, nil
		}
	}

	const trendingOrder = "(week_thread_count * 3 + week_like_count + week_repost_count * 2) DESC, start_time ASC"
	now := time.Now().UTC()

	var running []threadmodel.Event
	if err := r.readDB.
		Where("status = ?", "approved").
		Where("start_time <= ? AND end_time >= ?", now, now).
		Order(trendingOrder).
		Limit(limit).
		Find(&running).Error; err != nil {
		return nil, err
	}

	events = running
	// Fill the remaining slots with upcoming (future) events when fewer than
	// `limit` events are running today.
	if len(events) < limit {
		var future []threadmodel.Event
		if err := r.readDB.
			Where("status = ?", "approved").
			Where("start_time > ?", now).
			Order(trendingOrder).
			Limit(limit - len(events)).
			Find(&future).Error; err != nil {
			return nil, err
		}
		events = append(events, future...)
	}

	if r.cache != nil {
		if err := r.cache.Set(ctx, key, events, r.ttl); err == nil {
			r.cache.RegisterCacheKey(ctx, infrarepoShared.EventListsTag(), key)
		}
	}
	return events, nil
}

func (r *eventRepository) FindPending(limit int) ([]threadmodel.Event, error) {
	ctx := context.Background()
	key := eventPendingListKey(limit)
	var events []threadmodel.Event
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &events); err == nil {
			return events, nil
		}
	}

	if err := r.readDB.
		Where("status = ?", "pending").
		Order("created_at DESC").
		Limit(limit).
		Find(&events).Error; err != nil {
		return nil, err
	}

	if r.cache != nil {
		if err := r.cache.Set(ctx, key, events, r.ttl); err == nil {
			r.cache.RegisterCacheKey(ctx, infrarepoShared.EventListsTag(), key)
		}
	}
	return events, nil
}

func (r *eventRepository) FindAdmin(status, name string, page, limit int) ([]threadmodel.Event, int64, error) {
	q := r.readDB.Model(&threadmodel.Event{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if name != "" {
		q = q.Where("LOWER(title) LIKE ?", "%"+strings.ToLower(name)+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var events []threadmodel.Event
	offset := (page - 1) * limit
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&events).Error; err != nil {
		return nil, 0, err
	}
	return events, total, nil
}

func (r *eventRepository) FindByUserID(userID uint, status, name string, page, limit int) ([]threadmodel.Event, int64, error) {
	q := r.readDB.Model(&threadmodel.Event{}).Where("user_id = ?", userID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if name != "" {
		q = q.Where("LOWER(title) LIKE ?", "%"+strings.ToLower(name)+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var events []threadmodel.Event
	offset := (page - 1) * limit
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&events).Error; err != nil {
		return nil, 0, err
	}
	return events, total, nil
}

func (r *eventRepository) Update(event *threadmodel.Event) error {
	if err := r.writeDB.Model(event).UpdateColumns(map[string]any{
		"user_id":                event.UserID,
		"title":                  event.Title,
		"description":            event.Description,
		"location":               event.Location,
		"type":                   event.Type,
		"start_time":             event.StartTime,
		"end_time":               event.EndTime,
		"status":                 event.Status,
		"banner_object_name":     event.BannerObjectName,
		"attachment_object_name": event.AttachmentObjectName,
		"pdfs":                   event.PDFs,
		"updated_at":             gorm.Expr("NOW()"),
	}).Error; err != nil {
		return err
	}
	r.invalidateEvent(event.ID)
	r.invalidateEventThreadCaches(event.ID)
	return nil
}

func (r *eventRepository) UpdateStatus(id, status string) error {
	if err := r.writeDB.Model(&threadmodel.Event{}).
		Where("id = ?", id).
		UpdateColumns(map[string]any{
			"status":     status,
			"updated_at": gorm.Expr("NOW()"),
		}).Error; err != nil {
		return err
	}
	r.invalidateEvent(id)
	return nil
}

func (r *eventRepository) Delete(id string) error {
	r.invalidateEventThreadCaches(id)
	if err := r.writeDB.Delete(&threadmodel.Event{}, "id = ?", id).Error; err != nil {
		return err
	}
	r.invalidateEvent(id)
	return nil
}

func (r *eventRepository) FindThreads(eventID string, ownerUserID uint, requesterID *uint, filter, sort string, page, limit int) ([]threadmodel.Thread, int64, error) {
	_ = ownerUserID
	ctx := context.Background()
	key := eventThreadsKey(eventID, filter, sort, threadrepo.ThreadVisibilityCachePart(requesterID), page, limit)
	var result cachedEventThreads
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return result.Threads, result.Total, nil
		}
	}

	var threads []threadmodel.Thread
	var total int64
	threadInEvent := `EXISTS (
		SELECT 1 FROM thread_events te
		WHERE te.event_id = ?
		AND te.thread_id IN (
			threads.thread_id,
			COALESCE(threads.original_thread_id, 0),
			COALESCE(threads.root_id, 0),
			COALESCE(threads.parent_id, 0),
			COALESCE(original.thread_id, 0),
			COALESCE(original.root_id, 0),
			COALESCE(original.parent_id, 0)
		)
	)`
	threadTypes := []string{string(threadmodel.ThreadTypePost), string(threadmodel.ThreadTypeRepost), string(threadmodel.ThreadTypeReply)}
	q := threadrepo.ApplyThreadVisibility(r.readDB.Model(&threadmodel.Thread{}), requesterID).
		Joins("LEFT JOIN threads original ON original.thread_id = threads.original_thread_id").
		Joins("LEFT JOIN communities ON communities.community_id = COALESCE(threads.community_id, original.community_id)").
		Where(threadInEvent, eventID).
		Where("threads.type IN ?", threadTypes)
	if requesterID == nil {
		q = q.Where("COALESCE(threads.community_id, original.community_id) IS NULL OR communities.visibility = ?", communitymodel.CommunityVisibilityPublic)
	} else {
		q = q.
			Joins("LEFT JOIN community_members ON community_members.community_id = COALESCE(threads.community_id, original.community_id) AND community_members.user_id = ?", *requesterID).
			Where("COALESCE(threads.community_id, original.community_id) IS NULL OR communities.visibility = ? OR community_members.user_id = ?",
				communitymodel.CommunityVisibilityPublic, *requesterID)
	}
	switch filter {
	case "official":
		q = q.Where(`EXISTS (
			SELECT 1 FROM thread_events te2
			WHERE te2.event_id = ?
			AND te2.is_official = TRUE
			AND te2.thread_id IN (
				threads.thread_id,
				COALESCE(threads.original_thread_id, 0),
				COALESCE(threads.root_id, 0),
				COALESCE(threads.parent_id, 0),
				COALESCE(original.thread_id, 0),
				COALESCE(original.root_id, 0),
				COALESCE(original.parent_id, 0)
			)
		)`, eventID)
	case "unofficial":
		q = q.Where(`EXISTS (
			SELECT 1 FROM thread_events te2
			WHERE te2.event_id = ?
			AND te2.is_official = FALSE
			AND te2.thread_id IN (
				threads.thread_id,
				COALESCE(threads.original_thread_id, 0),
				COALESCE(threads.root_id, 0),
				COALESCE(threads.parent_id, 0),
				COALESCE(original.thread_id, 0),
				COALESCE(original.root_id, 0),
				COALESCE(original.parent_id, 0)
			)
		)`, eventID)
	}
	q.Count(&total)
	offset := (page - 1) * limit
	if err := threadrepo.PreloadThreadDetails(q).
		Order(eventThreadSortClause(sort)).
		Offset(offset).Limit(limit).
		Find(&threads).Error; err != nil {
		return nil, 0, err
	}
	result = cachedEventThreads{Threads: threads, Total: total}
	if r.cache != nil {
		if err := r.cache.Set(ctx, key, result, threadrepo.ThreadFeedTTL); err == nil {
			r.cache.RegisterCacheKey(ctx, infrarepoShared.GlobalThreadFeedsTag(), key)
			r.cache.RegisterCacheKey(ctx, infrarepoShared.EventThreadsGlobalTag(), key)
			r.cache.RegisterCacheKey(ctx, infrarepoShared.EventThreadsByEventTag(eventID), key)
			if requesterID != nil && *requesterID != 0 {
				r.cache.RegisterCacheKey(ctx, infrarepoShared.ViewerThreadFeedsTag(*requesterID), key)
			}
			for _, t := range result.Threads {
				if t.ThreadID != 0 {
					r.cache.RegisterCacheKey(ctx, infrarepoShared.ThreadFeedsTag(t.ThreadID), key)
				}
			}
		}
	}
	return result.Threads, result.Total, nil
}

func (r *eventRepository) IsThreadInEvent(eventID string, threadID uint) (bool, error) {
	var count int64
	err := r.readDB.Raw(`
		SELECT COUNT(*) FROM threads t
		WHERE t.thread_id = ?
		AND EXISTS (
			SELECT 1 FROM thread_events te
			WHERE te.event_id = ?
			AND te.thread_id IN (t.thread_id, COALESCE(t.original_thread_id, 0), COALESCE(t.root_id, 0), COALESCE(t.parent_id, 0))
		)
	`, threadID, eventID).Scan(&count).Error
	return count > 0, err
}

func (r *eventRepository) PatchEventCounterByThreadID(ctx context.Context, threadID uint, column string, delta int) error {
	if delta == 0 {
		return nil
	}

	var event threadmodel.ThreadEvent
	if err := r.writeDB.WithContext(ctx).Select("event_id").Where("thread_id = ?", threadID).First(&event).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		var thread threadmodel.Thread
		if err := r.writeDB.WithContext(ctx).Select("root_id", "parent_id", "original_thread_id").Where("thread_id = ?", threadID).First(&thread).Error; err != nil {
			return err
		}
		if thread.RootID != nil {
			threadID = *thread.RootID
		} else if thread.ParentID != nil {
			threadID = *thread.ParentID
		} else if thread.OriginalThreadID != nil {
			threadID = *thread.OriginalThreadID
		} else {
			return nil
		}
		if err := r.writeDB.WithContext(ctx).Select("event_id").Where("thread_id = ?", threadID).First(&event).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
	}
	if event.EventID == "" {
		return nil
	}

	updates := map[string]any{
		column: gorm.Expr("GREATEST("+column+" + ?, 0)", delta),
	}
	switch column {
	case "thread_count":
		updates["week_thread_count"] = gorm.Expr("GREATEST(week_thread_count + ?, 0)", delta)
	case "like_count":
		updates["week_like_count"] = gorm.Expr("GREATEST(week_like_count + ?, 0)", delta)
	case "repost_count":
		updates["week_repost_count"] = gorm.Expr("GREATEST(week_repost_count + ?, 0)", delta)
	case "week_thread_count", "week_like_count", "week_repost_count":
	default:
		return nil
	}

	if err := r.writeDB.WithContext(ctx).Model(&threadmodel.Event{}).
		Where("id = ?", event.EventID).
		UpdateColumns(updates).Error; err != nil {
		return err
	}
	infrarepoShared.InvalidateEventCounterCaches(r.cache, event.EventID)
	return nil
}

func (r *eventRepository) invalidateLists() {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	// infrarepoShared.EventListsTag covers event:list/approved:list/pending:list at every Set
	// site (FindGrouped, FindGroupedApproved, FindPending).
	_ = r.cache.InvalidateTag(ctx, infrarepoShared.EventListsTag())
}

func (r *eventRepository) invalidateEvent(id string) {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.Delete(ctx, infrarepoShared.EventIDKey(id))
	// infrarepoShared.EventListsTag → all event list caches; infrarepoShared.EventThreadsByEventTag → all
	// cached event:threads:<id>:* pages for this event.
	_ = r.cache.InvalidateTags(ctx, infrarepoShared.EventListsTag(), infrarepoShared.EventThreadsByEventTag(id))
}

func (r *eventRepository) invalidateEventThreadCaches(id string) {
	if r.cache == nil {
		return
	}

	var threads []threadmodel.Thread
	if err := r.readDB.
		Table("threads").
		Select("DISTINCT threads.thread_id, threads.user_id, threads.root_id, threads.community_id").
		Joins("LEFT JOIN thread_events ON thread_events.thread_id = threads.thread_id").
		Where("thread_events.event_id = ? OR threads.original_thread_id IN (?)",
			id,
			r.readDB.Table("thread_events").Select("thread_id").Where("event_id = ?", id),
		).
		Scan(&threads).Error; err != nil {
		return
	}

	ctx := context.Background()
	userIDSet := make(map[uint]struct{}, len(threads))
	treeIDSet := make(map[uint]struct{}, len(threads))
	communityIDSet := make(map[uint]struct{})
	threadIDs := make([]uint, 0, len(threads))
	for _, thread := range threads {
		threadIDs = append(threadIDs, thread.ThreadID)
		_ = r.cache.Delete(ctx, threadrepo.ThreadKey(thread.ThreadID), threadrepo.ParentThreadKey(thread.ThreadID))
		_ = r.cache.InvalidateTags(ctx,
			infrarepoShared.ThreadFeedsTag(thread.ThreadID),
			infrarepoShared.InteractionThreadCacheKeysKey("likes", thread.ThreadID),
			infrarepoShared.InteractionThreadCacheKeysKey("bookmarks", thread.ThreadID),
			infrarepoShared.InteractionThreadCacheKeysKey("reposts", thread.ThreadID),
		)

		if thread.UserID != 0 {
			userIDSet[thread.UserID] = struct{}{}
		}
		treeID := thread.ThreadID
		if thread.RootID != nil && *thread.RootID != 0 {
			treeID = *thread.RootID
		}
		if treeID != 0 {
			treeIDSet[treeID] = struct{}{}
		}
		if thread.CommunityID != nil && *thread.CommunityID != 0 {
			communityIDSet[*thread.CommunityID] = struct{}{}
		}
	}
	// Per-tree, per-community, per-user, per-follower → reverse-index tags.
	// Each tag set was populated when the corresponding cache was first
	// written (storeFeedIDs / storeChainFeedIDs / findUserFeed / etc).
	chainTags := make([]string, 0, len(treeIDSet))
	for treeID := range treeIDSet {
		chainTags = append(chainTags, infrarepoShared.ThreadChainTag(treeID))
	}
	communityTags := make([]string, 0, len(communityIDSet))
	for communityID := range communityIDSet {
		communityTags = append(communityTags, infrarepoShared.ThreadCommunityTag(communityID))
	}
	userTags := make([]string, 0, len(userIDSet))
	userIDs := make([]uint, 0, len(userIDSet))
	for userID := range userIDSet {
		userIDs = append(userIDs, userID)
		userTags = append(userTags, infrarepoShared.UserFeedsTag(userID))
	}
	if len(chainTags) > 0 {
		_ = r.cache.InvalidateTags(ctx, chainTags...)
	}
	if len(communityTags) > 0 {
		_ = r.cache.InvalidateTags(ctx, communityTags...)
	}
	if len(userTags) > 0 {
		_ = r.cache.InvalidateTags(ctx, userTags...)
	}
	if len(userIDs) > 0 {
		var followerIDs []uint
		if err := r.readDB.Table("user_relationships").
			Distinct("follower_id").
			Where("following_id IN ?", userIDs).
			Pluck("follower_id", &followerIDs).Error; err == nil {
			followerTags := make([]string, 0, len(followerIDs))
			for _, followerID := range followerIDs {
				followerTags = append(followerTags, infrarepoShared.UserFeedsTag(followerID))
			}
			if len(followerTags) > 0 {
				_ = r.cache.InvalidateTags(ctx, followerTags...)
			}
		}
	}
	r.invalidateInteractionUserCaches(ctx, "likes", threadIDs)
	r.invalidateInteractionUserCaches(ctx, "bookmarks", threadIDs)
	r.invalidateInteractionUserCaches(ctx, "reposts", threadIDs)
	_ = r.cache.Delete(ctx, "feed:trending:global")
	// infrarepoShared.GlobalThreadFeedsTag covers thread:roots/all-communities/my-communities,
	// community:threads:top/latest at every Set site. feed:foryou:* is still
	// global SCAN — no per-user tag covers global wipe; mutations on a single
	// event affect cached foryou candidates for every viewer.
	_ = r.cache.InvalidateTag(ctx, infrarepoShared.GlobalThreadFeedsTag())
	_ = r.cache.DeleteByPattern(ctx, "feed:foryou:*")
	infrarepoShared.InvalidateExploreCache(ctx, r.cache)
}

func (r *eventRepository) invalidateInteractionUserCaches(ctx context.Context, table string, threadIDs []uint) {
	if len(threadIDs) == 0 {
		return
	}
	var userIDs []uint
	if err := r.readDB.Table(table).
		Distinct("user_id").
		Where("thread_id IN ?", threadIDs).
		Pluck("user_id", &userIDs).Error; err != nil {
		return
	}
	if len(userIDs) == 0 {
		return
	}
	tags := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		tags = append(tags, infrarepoShared.InteractionUserListTag(table, userID))
	}
	_ = r.cache.InvalidateTags(ctx, tags...)
}
