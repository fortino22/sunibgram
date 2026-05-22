package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
)

type eventRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
	ttl     time.Duration
}

type cachedEvents struct {
	Trending []model.Event
	Running  []model.Event
	Future   []model.Event
}

type cachedEventThreads struct {
	Threads []model.Thread
	Total   int64
}

func NewEventRepository(writeDB, readDB *gorm.DB, c *cache.Cache) domainrepo.IEventRepository {
	return &eventRepository{writeDB: writeDB, readDB: readDB, cache: c, ttl: cache.DefaultTTL}
}

func eventIDKey(id string) string {
	return fmt.Sprintf("event:id:%s", id)
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

func (r *eventRepository) Save(event *model.Event) error {
	if err := r.writeDB.Create(event).Error; err != nil {
		return err
	}
	r.invalidateLists()
	return nil
}

func (r *eventRepository) FindByID(id string) (*model.Event, error) {
	ctx := context.Background()
	key := eventIDKey(id)
	var event model.Event
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

func (r *eventRepository) FindGrouped(limit int) ([]model.Event, []model.Event, []model.Event, error) {
	ctx := context.Background()
	key := eventListKey(limit)
	var result cachedEvents
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return result.Trending, result.Running, result.Future, nil
		}
	}

	var trending []model.Event
	if err := r.readDB.
		Order("(week_thread_count * 3 + week_like_count + week_repost_count * 2) DESC, start_time ASC").
		Limit(limit).
		Find(&trending).Error; err != nil {
		return nil, nil, nil, err
	}

	now := time.Now().UTC()
	var running []model.Event
	if err := r.readDB.
		Where("start_time <= ? AND end_time >= ?", now, now).
		Order("start_time ASC").
		Limit(limit).
		Find(&running).Error; err != nil {
		return nil, nil, nil, err
	}

	var future []model.Event
	if err := r.readDB.
		Where("start_time > ?", now).
		Order("start_time ASC").
		Limit(limit).
		Find(&future).Error; err != nil {
		return nil, nil, nil, err
	}

	result = cachedEvents{Trending: trending, Running: running, Future: future}
	if r.cache != nil {
		_ = r.cache.Set(ctx, key, result, r.ttl)
	}
	return result.Trending, result.Running, result.Future, nil
}

func (r *eventRepository) FindGroupedApproved(limit int) ([]model.Event, []model.Event, []model.Event, error) {
	ctx := context.Background()
	key := eventApprovedListKey(limit)
	var result cachedEvents
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return result.Trending, result.Running, result.Future, nil
		}
	}

	var trending []model.Event
	if err := r.readDB.
		Where("status = ?", "approved").
		Order("(week_thread_count * 3 + week_like_count + week_repost_count * 2) DESC, start_time ASC").
		Limit(limit).
		Find(&trending).Error; err != nil {
		return nil, nil, nil, err
	}

	now := time.Now().UTC()
	var running []model.Event
	if err := r.readDB.
		Where("status = ?", "approved").
		Where("start_time <= ? AND end_time >= ?", now, now).
		Order("start_time ASC").
		Limit(limit).
		Find(&running).Error; err != nil {
		return nil, nil, nil, err
	}

	var future []model.Event
	if err := r.readDB.
		Where("status = ?", "approved").
		Where("start_time > ?", now).
		Order("start_time ASC").
		Limit(limit).
		Find(&future).Error; err != nil {
		return nil, nil, nil, err
	}

	result = cachedEvents{Trending: trending, Running: running, Future: future}
	if r.cache != nil {
		_ = r.cache.Set(ctx, key, result, r.ttl)
	}
	return result.Trending, result.Running, result.Future, nil
}

func (r *eventRepository) FindPending(limit int) ([]model.Event, error) {
	ctx := context.Background()
	key := eventPendingListKey(limit)
	var events []model.Event
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
		_ = r.cache.Set(ctx, key, events, r.ttl)
	}
	return events, nil
}

func (r *eventRepository) FindAdmin(status, name string, page, limit int) ([]model.Event, int64, error) {
	q := r.readDB.Model(&model.Event{})
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

	var events []model.Event
	offset := (page - 1) * limit
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&events).Error; err != nil {
		return nil, 0, err
	}
	return events, total, nil
}

func (r *eventRepository) FindByUserID(userID uint, status string, page, limit int) ([]model.Event, int64, error) {
	q := r.readDB.Model(&model.Event{}).Where("user_id = ?", userID)
	if status != "" {
		q = q.Where("status = ?", status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var events []model.Event
	offset := (page - 1) * limit
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&events).Error; err != nil {
		return nil, 0, err
	}
	return events, total, nil
}

func (r *eventRepository) Update(event *model.Event) error {
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
		"banner_bucket":          event.BannerBucket,
		"attachment_object_name": event.AttachmentObjectName,
		"attachment_bucket":      event.AttachmentBucket,
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
	if err := r.writeDB.Model(&model.Event{}).
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
	if err := r.writeDB.Delete(&model.Event{}, "id = ?", id).Error; err != nil {
		return err
	}
	r.invalidateEvent(id)
	return nil
}

func (r *eventRepository) FindThreads(eventID string, ownerUserID uint, requesterID *uint, filter, sort string, page, limit int) ([]model.Thread, int64, error) {
	ctx := context.Background()
	key := eventThreadsKey(eventID, filter, sort, threadVisibilityCachePart(requesterID), page, limit)
	var result cachedEventThreads
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return result.Threads, result.Total, nil
		}
	}

	var threads []model.Thread
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
	threadTypes := []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost), string(model.ThreadTypeReply)}
	q := applyThreadVisibility(r.readDB.Model(&model.Thread{}), requesterID).
		Joins("LEFT JOIN threads original ON original.thread_id = threads.original_thread_id").
		Joins("LEFT JOIN communities ON communities.community_id = COALESCE(threads.community_id, original.community_id)").
		Where(threadInEvent, eventID).
		Where("threads.type IN ?", threadTypes)
	if requesterID == nil {
		q = q.Where("COALESCE(threads.community_id, original.community_id) IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic)
	} else {
		q = q.
			Joins("LEFT JOIN community_members ON community_members.community_id = COALESCE(threads.community_id, original.community_id) AND community_members.user_id = ?", *requesterID).
			Where("COALESCE(threads.community_id, original.community_id) IS NULL OR communities.visibility = ? OR community_members.user_id = ?",
				model.CommunityVisibilityPublic, *requesterID)
	}
	switch filter {
	case "official":
		q = q.Where("threads.user_id = ?", ownerUserID)
	case "unofficial":
		q = q.Where("threads.user_id != ?", ownerUserID)
	}
	q.Count(&total)
	offset := (page - 1) * limit
	if err := preloadThreadDetails(q).
		Order(eventThreadSortClause(sort)).
		Offset(offset).Limit(limit).
		Find(&threads).Error; err != nil {
		return nil, 0, err
	}
	result = cachedEventThreads{Threads: threads, Total: total}
	if r.cache != nil {
		if err := r.cache.Set(ctx, key, result, threadFeedTTL); err == nil {
			for _, t := range result.Threads {
				if t.ThreadID != 0 {
					r.cache.RegisterCacheKey(ctx, threadFeedsTag(t.ThreadID), key)
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

	var event model.ThreadEvent
	if err := r.writeDB.WithContext(ctx).Select("event_id").Where("thread_id = ?", threadID).First(&event).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		var thread model.Thread
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

	if err := r.writeDB.WithContext(ctx).Model(&model.Event{}).
		Where("id = ?", event.EventID).
		UpdateColumns(updates).Error; err != nil {
		return err
	}
	invalidateEventCounterCaches(r.cache, event.EventID)
	return nil
}

func (r *eventRepository) invalidateLists() {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.DeleteByPatterns(ctx,
		"event:list:*",
		"event:approved:list:*",
		"event:pending:list:*",
	)
}

func (r *eventRepository) invalidateEvent(id string) {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.Delete(ctx, eventIDKey(id))
	_ = r.cache.DeleteByPatterns(ctx,
		"event:list:*",
		"event:approved:list:*",
		"event:pending:list:*",
		fmt.Sprintf("event:threads:%s:*", id),
	)
}

func (r *eventRepository) invalidateEventThreadCaches(id string) {
	if r.cache == nil {
		return
	}

	var threads []model.Thread
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
		_ = r.cache.Delete(ctx, threadKey(thread.ThreadID), parentThreadKey(thread.ThreadID))
		_ = r.cache.InvalidateTags(ctx,
			threadFeedsTag(thread.ThreadID),
			interactionThreadCacheKeysKey("likes", thread.ThreadID),
			interactionThreadCacheKeysKey("bookmarks", thread.ThreadID),
			interactionThreadCacheKeysKey("reposts", thread.ThreadID),
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
	for treeID := range treeIDSet {
		_ = r.cache.DeleteByPattern(ctx, fmt.Sprintf("thread:chain:%d:*", treeID))
	}
	for communityID := range communityIDSet {
		_ = r.cache.DeleteByPattern(ctx, fmt.Sprintf("thread:community:%d:*", communityID))
	}
	userIDs := make([]uint, 0, len(userIDSet))
	for userID := range userIDSet {
		userIDs = append(userIDs, userID)
		_ = r.cache.DeleteByPattern(ctx, fmt.Sprintf("thread:userfeed:%d:*", userID))
		_ = r.cache.DeleteByPattern(ctx, fmt.Sprintf("thread:usermedia:%d:*", userID))
	}
	if len(userIDs) > 0 {
		var followerIDs []uint
		if err := r.readDB.Table("user_relationships").
			Distinct("follower_id").
			Where("following_id IN ?", userIDs).
			Pluck("follower_id", &followerIDs).Error; err == nil {
			for _, followerID := range followerIDs {
				_ = r.cache.DeleteByPattern(ctx, fmt.Sprintf("thread:following:%d:*", followerID))
			}
		}
	}
	r.invalidateInteractionUserCaches(ctx, "likes", threadIDs)
	r.invalidateInteractionUserCaches(ctx, "bookmarks", threadIDs)
	r.invalidateInteractionUserCaches(ctx, "reposts", threadIDs)
	_ = r.cache.Delete(ctx, "feed:trending:global")
	_ = r.cache.DeleteByPatterns(ctx,
		"feed:foryou:*",
		"thread:roots:*",
		"thread:all-communities:*",
		"thread:my-communities:*",
		"community:threads:top:*",
		"community:threads:latest:*",
	)
	invalidateExploreCache(ctx, r.cache)
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
	for _, userID := range userIDs {
		_ = r.cache.DeleteByPattern(ctx, fmt.Sprintf("%s:user:%d:*", table, userID))
	}
}

func interactionThreadCacheKeysKey(table string, threadID uint) string {
	return fmt.Sprintf("%s:thread:%d:cache_keys", table, threadID)
}
