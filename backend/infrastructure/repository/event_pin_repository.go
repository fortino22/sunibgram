package repository

import (
	"context"
	"fmt"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
)

type EventPinRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewEventPinRepository(writeDB, readDB *gorm.DB, c *cache.Cache) domainrepo.IEventPinRepository {
	return &EventPinRepository{writeDB: writeDB, readDB: readDB, cache: c}
}

type cachedEventPins struct {
	Events []model.Event
	Total  int64
}

func eventPinListKey(userID uint, page, limit int) string {
	return fmt.Sprintf("event:pins:user:%d:page:%d:limit:%d", userID, page, limit)
}

func (r *EventPinRepository) Pin(pin *model.UserEventPin) error {
	if err := r.writeDB.Create(pin).Error; err != nil {
		return err
	}
	r.InvalidateUser(pin.UserID)
	return nil
}

func (r *EventPinRepository) Unpin(userID uint, eventID string) error {
	if err := r.writeDB.
		Where("user_id = ? AND event_id = ?", userID, eventID).
		Delete(&model.UserEventPin{}).Error; err != nil {
		return err
	}
	r.InvalidateUser(userID)
	return nil
}

func (r *EventPinRepository) Exists(userID uint, eventID string) (bool, error) {
	var count int64
	err := r.readDB.Model(&model.UserEventPin{}).
		Where("user_id = ? AND event_id = ?", userID, eventID).
		Count(&count).Error
	return count > 0, err
}

func (r *EventPinRepository) CountByUser(userID uint) (int64, error) {
	var count int64
	err := r.readDB.Model(&model.UserEventPin{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

func (r *EventPinRepository) FindMyPins(userID uint, page, limit int) ([]model.Event, int64, error) {
	ctx := context.Background()
	key := eventPinListKey(userID, page, limit)
	var result cachedEventPins
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return result.Events, result.Total, nil
		}
	}

	var total int64
	r.readDB.Model(&model.Event{}).
		Joins("JOIN user_event_pins ON user_event_pins.event_id = events.id").
		Where("user_event_pins.user_id = ?", userID).
		Count(&total)

	var events []model.Event
	offset := (page - 1) * limit
	if err := r.readDB.
		Joins("JOIN user_event_pins ON user_event_pins.event_id = events.id").
		Where("user_event_pins.user_id = ?", userID).
		Order("user_event_pins.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&events).Error; err != nil {
		return nil, 0, err
	}

	if r.cache != nil {
		result = cachedEventPins{Events: events, Total: total}
		if err := r.cache.Set(ctx, key, result, cache.DefaultTTL); err == nil {
			r.cache.RegisterCacheKey(ctx, eventPinsUserTag(userID), key)
		}
	}
	return events, total, nil
}

func (r *EventPinRepository) InvalidateUser(userID uint) {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.InvalidateTag(ctx, eventPinsUserTag(userID))
}

func (r *EventPinRepository) InvalidateByEvent(eventID string) {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	var userIDs []uint
	if err := r.readDB.Model(&model.UserEventPin{}).
		Where("event_id = ?", eventID).
		Distinct("user_id").
		Pluck("user_id", &userIDs).Error; err != nil {
		return
	}
	if len(userIDs) == 0 {
		return
	}
	tags := make([]string, 0, len(userIDs))
	for _, uid := range userIDs {
		tags = append(tags, eventPinsUserTag(uid))
	}
	_ = r.cache.InvalidateTags(ctx, tags...)
}
