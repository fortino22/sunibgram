package repository

import (
	"context"
	"errors"
	"fmt"
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

func eventThreadsKey(id string, page, limit int) string {
	return fmt.Sprintf("event:threads:%s:page:%d:limit:%d", id, page, limit)
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

func (r *eventRepository) Update(event *model.Event) error {
	if err := r.writeDB.Model(event).UpdateColumns(map[string]any{
		"title":              event.Title,
		"description":        event.Description,
		"location":           event.Location,
		"start_time":         event.StartTime,
		"end_time":           event.EndTime,
		"banner_object_name": event.BannerObjectName,
		"banner_bucket":      event.BannerBucket,
		"updated_at":         gorm.Expr("NOW()"),
	}).Error; err != nil {
		return err
	}
	r.invalidateEvent(event.ID)
	r.invalidateEventThreadCaches(event.ID)
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

func (r *eventRepository) FindThreads(eventID string, page, limit int) ([]model.Thread, int64, error) {
	ctx := context.Background()
	key := eventThreadsKey(eventID, page, limit)
	var result cachedEventThreads
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return result.Threads, result.Total, nil
		}
	}

	var threads []model.Thread
	var total int64
	q := r.readDB.Model(&model.Thread{}).
		Joins("JOIN thread_events ON thread_events.thread_id = threads.thread_id").
		Where("thread_events.event_id = ?", eventID)
	q.Count(&total)
	offset := (page - 1) * limit
	if err := preloadThreadDetails(q).
		Order("thread_events.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&threads).Error; err != nil {
		return nil, 0, err
	}
	result = cachedEventThreads{Threads: threads, Total: total}
	if r.cache != nil {
		_ = r.cache.Set(ctx, key, result, r.ttl)
	}
	return result.Threads, result.Total, nil
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
		if err := r.writeDB.WithContext(ctx).Select("root_id", "parent_id").First(&thread, threadID).Error; err != nil {
			return err
		}
		if thread.RootID != nil {
			threadID = *thread.RootID
		} else if thread.ParentID != nil {
			threadID = *thread.ParentID
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
	_ = r.cache.DeleteByPattern(context.Background(), "event:list:*")
}

func (r *eventRepository) invalidateEvent(id string) {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.Delete(ctx, eventIDKey(id))
	_ = r.cache.DeleteByPattern(ctx, "event:list:*")
	_ = r.cache.DeleteByPattern(ctx, fmt.Sprintf("event:threads:%s:*", id))
}

func (r *eventRepository) invalidateEventThreadCaches(id string) {
	if r.cache == nil {
		return
	}

	var threadIDs []uint
	if err := r.readDB.
		Table("threads").
		Select("threads.thread_id").
		Joins("LEFT JOIN thread_events ON thread_events.thread_id = threads.thread_id").
		Where("thread_events.event_id = ? OR threads.original_thread_id IN (?)",
			id,
			r.readDB.Table("thread_events").Select("thread_id").Where("event_id = ?", id),
		).
		Pluck("threads.thread_id", &threadIDs).Error; err != nil {
		return
	}

	ctx := context.Background()
	for _, threadID := range threadIDs {
		_ = r.cache.Delete(ctx, threadKey(threadID))
		_ = r.cache.Delete(ctx, threadFeedKeysKey(threadID))
	}
	_ = r.cache.DeleteByPattern(ctx, "thread:roots:*")
	_ = r.cache.DeleteByPattern(ctx, "thread:chain:*")
	_ = r.cache.DeleteByPattern(ctx, "thread:community:*")
	_ = r.cache.DeleteByPattern(ctx, "thread:userfeed:*")
	_ = r.cache.DeleteByPattern(ctx, "thread:usermedia:*")
	_ = r.cache.DeleteByPattern(ctx, "thread:following:*")
	_ = r.cache.DeleteByPattern(ctx, "likes:user:*")
	_ = r.cache.DeleteByPattern(ctx, "bookmarks:user:*")
	_ = r.cache.DeleteByPattern(ctx, "reposts:user:*")
	invalidateExploreCache(ctx, r.cache)
}
