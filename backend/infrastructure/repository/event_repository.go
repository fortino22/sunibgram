package repository

import (
	"context"
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
	Events []model.Event
	Total  int64
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

func eventListKey(page, limit int) string {
	return fmt.Sprintf("event:list:page:%d:limit:%d", page, limit)
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

func (r *eventRepository) FindAll(page, limit int) ([]model.Event, int64, error) {
	ctx := context.Background()
	key := eventListKey(page, limit)
	var result cachedEvents
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return result.Events, result.Total, nil
		}
	}

	var events []model.Event
	var total int64
	r.readDB.Model(&model.Event{}).Count(&total)
	offset := (page - 1) * limit
	if err := r.readDB.Order("start_time ASC").Offset(offset).Limit(limit).Find(&events).Error; err != nil {
		return nil, 0, err
	}
	result = cachedEvents{Events: events, Total: total}
	if r.cache != nil {
		_ = r.cache.Set(ctx, key, result, r.ttl)
	}
	return result.Events, result.Total, nil
}

func (r *eventRepository) Update(event *model.Event) error {
	if err := r.writeDB.Model(event).UpdateColumns(map[string]any{
		"title":       event.Title,
		"description": event.Description,
		"location":    event.Location,
		"start_time":  event.StartTime,
		"end_time":    event.EndTime,
		"updated_at":  gorm.Expr("NOW()"),
	}).Error; err != nil {
		return err
	}
	r.invalidateEvent(event.ID)
	return nil
}

func (r *eventRepository) Delete(id string) error {
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
	if err := q.Preload("User", preloadThreadUser).
		Preload("OriginalThread.User", preloadThreadUser).
		Preload("Community").
		Preload("OriginalThread.Community").
		Preload("User.UserFaculty.Faculty").
		Preload("OriginalThread.User.UserFaculty.Faculty").
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
