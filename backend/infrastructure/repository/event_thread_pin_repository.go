package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	eventthreadpinmodel "github.com/sunibgram/domain/model/event_thread_pin"
	eventthreadpinrepoiface "github.com/sunibgram/domain/repository/event_thread_pin"
	"github.com/sunibgram/infrastructure/cache"
)

type EventThreadPinRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewEventThreadPinRepository(writeDB, readDB *gorm.DB, c *cache.Cache) eventthreadpinrepoiface.IEventThreadPinRepository {
	return &EventThreadPinRepository{writeDB: writeDB, readDB: readDB, cache: c}
}

func eventThreadPinKey(eventID string) string {
	return fmt.Sprintf("event:thread_pin:%s", eventID)
}

func (r *EventThreadPinRepository) Pin(pin *eventthreadpinmodel.EventThreadPin) error {
	if err := r.writeDB.Create(pin).Error; err != nil {
		return err
	}
	r.invalidate(pin.EventID)
	return nil
}

func (r *EventThreadPinRepository) Unpin(eventID string, threadID uint) error {
	if err := r.writeDB.Where("event_id = ? AND thread_id = ?", eventID, threadID).
		Delete(&eventthreadpinmodel.EventThreadPin{}).Error; err != nil {
		return err
	}
	r.invalidate(eventID)
	return nil
}

func (r *EventThreadPinRepository) Exists(eventID string, threadID uint) (bool, error) {
	var count int64
	err := r.readDB.Model(&eventthreadpinmodel.EventThreadPin{}).
		Where("event_id = ? AND thread_id = ?", eventID, threadID).
		Count(&count).Error
	return count > 0, err
}

func (r *EventThreadPinRepository) FindByEventID(eventID string) (*eventthreadpinmodel.EventThreadPin, error) {
	ctx := context.Background()
	var pin eventthreadpinmodel.EventThreadPin
	if r.cache != nil {
		if err := r.cache.Get(ctx, eventThreadPinKey(eventID), &pin); err == nil {
			if pin.ThreadID == 0 {
				return nil, nil
			}
			return &pin, nil
		}
	}
	err := r.readDB.Where("event_id = ?", eventID).First(&pin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if r.cache != nil {
			r.cache.Set(ctx, eventThreadPinKey(eventID), eventthreadpinmodel.EventThreadPin{EventID: eventID}, cache.DefaultTTL)
		}
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if r.cache != nil {
		r.cache.Set(ctx, eventThreadPinKey(eventID), pin, cache.DefaultTTL)
	}
	return &pin, nil
}

func (r *EventThreadPinRepository) invalidate(eventID string) {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.Delete(ctx, eventThreadPinKey(eventID))
	_ = r.cache.InvalidateTag(ctx, eventThreadsByEventTag(eventID))
}
