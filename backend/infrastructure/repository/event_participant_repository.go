package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	eventparticipantmodel "github.com/sunibgram/domain/model/event_participant"
	threadmodel "github.com/sunibgram/domain/model/thread"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
)

type EventParticipantRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewEventParticipantRepository(writeDB, readDB *gorm.DB, c *cache.Cache) domainrepo.IEventParticipantRepository {
	return &EventParticipantRepository{writeDB: writeDB, readDB: readDB, cache: c}
}

type cachedJoinedEvents struct {
	Events []threadmodel.Event
	Total  int64
}

func eventJoinedListKey(userID uint, page, limit int) string {
	return fmt.Sprintf("event:joined:user:%d:page:%d:limit:%d", userID, page, limit)
}

func (r *EventParticipantRepository) Join(p *eventparticipantmodel.EventParticipant) error {
	if err := r.writeDB.Create(p).Error; err != nil {
		return err
	}
	r.InvalidateUser(p.UserID)
	return nil
}

func (r *EventParticipantRepository) Leave(userID uint, eventID string) error {
	if err := r.writeDB.
		Where("user_id = ? AND event_id = ?", userID, eventID).
		Delete(&eventparticipantmodel.EventParticipant{}).Error; err != nil {
		return err
	}
	r.InvalidateUser(userID)
	return nil
}

func (r *EventParticipantRepository) Exists(userID uint, eventID string) (bool, error) {
	var count int64
	err := r.readDB.Model(&eventparticipantmodel.EventParticipant{}).
		Where("user_id = ? AND event_id = ?", userID, eventID).
		Count(&count).Error
	return count > 0, err
}

func (r *EventParticipantRepository) FindMyJoined(userID uint, page, limit int) ([]threadmodel.Event, int64, error) {
	ctx := context.Background()
	key := eventJoinedListKey(userID, page, limit)
	var result cachedJoinedEvents
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return result.Events, result.Total, nil
		}
	}

	var total int64
	r.readDB.Model(&threadmodel.Event{}).
		Joins("JOIN event_participants ON event_participants.event_id = events.id").
		Where("event_participants.user_id = ?", userID).
		Count(&total)

	var events []threadmodel.Event
	offset := (page - 1) * limit
	if err := r.readDB.
		Joins("JOIN event_participants ON event_participants.event_id = events.id").
		Where("event_participants.user_id = ?", userID).
		Order("event_participants.joined_at DESC").
		Offset(offset).Limit(limit).
		Find(&events).Error; err != nil {
		return nil, 0, err
	}

	if r.cache != nil {
		result = cachedJoinedEvents{Events: events, Total: total}
		if err := r.cache.Set(ctx, key, result, cache.DefaultTTL); err == nil {
			r.cache.RegisterCacheKey(ctx, eventJoinedUserTag(userID), key)
		}
	}
	return events, total, nil
}

func (r *EventParticipantRepository) InvalidateUser(userID uint) {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.InvalidateTag(ctx, eventJoinedUserTag(userID))
}

func (r *EventParticipantRepository) InvalidateByEvent(eventID string) {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	var userIDs []uint
	if err := r.readDB.Model(&eventparticipantmodel.EventParticipant{}).
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
		tags = append(tags, eventJoinedUserTag(uid))
	}
	_ = r.cache.InvalidateTags(ctx, tags...)
}
