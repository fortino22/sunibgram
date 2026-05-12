package repository

import (
	"context"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
)

func eventIDForThread(db *gorm.DB, threadID uint) (string, bool) {
	var event model.ThreadEvent
	if err := db.Select("event_id").Where("thread_id = ?", threadID).First(&event).Error; err != nil {
		return "", false
	}
	return event.EventID, event.EventID != ""
}

func patchEventCounter(db *gorm.DB, c *cache.Cache, eventID, column string, delta int) {
	if eventID == "" || delta == 0 {
		return
	}
	db.Model(&model.Event{}).
		Where("id = ?", eventID).
		UpdateColumn(column, gorm.Expr("GREATEST("+column+" + ?, 0)", delta))
	invalidateEventCounterCaches(c, eventID)
}

func patchEventCounterByThreadID(db *gorm.DB, c *cache.Cache, threadID uint, column string, delta int) {
	if eventID, ok := eventIDForThread(db, threadID); ok {
		patchEventCounter(db, c, eventID, column, delta)
	}
}

func invalidateEventCounterCaches(c *cache.Cache, eventID string) {
	if c == nil {
		return
	}
	ctx := context.Background()
	_ = c.Delete(ctx, eventIDKey(eventID))
	_ = c.DeleteByPattern(ctx, "event:list:*")
	_ = c.DeleteByPattern(ctx, "event:threads:"+eventID+":*")
}
