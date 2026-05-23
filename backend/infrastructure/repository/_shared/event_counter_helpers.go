package infrarepoShared

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	threadmodel "github.com/sunibgram/domain/model/thread"
	"github.com/sunibgram/infrastructure/cache"
)

func EventIDForThread(db *gorm.DB, threadID uint) (string, bool) {
	var event threadmodel.ThreadEvent
	if err := db.Select("event_id").Where("thread_id = ?", threadID).First(&event).Error; err != nil {
		return "", false
	}
	return event.EventID, event.EventID != ""
}

func PatchEventCounter(db *gorm.DB, c *cache.Cache, eventID, column string, delta int) {
	if eventID == "" || delta == 0 {
		return
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
	}
	db.Model(&threadmodel.Event{}).
		Where("id = ?", eventID).
		UpdateColumns(updates)
	InvalidateEventCounterCaches(c, eventID)
}

func PatchEventCounters(db *gorm.DB, c *cache.Cache, eventID string, threadDelta, likeDelta, repostDelta int) {
	if eventID == "" || (threadDelta == 0 && likeDelta == 0 && repostDelta == 0) {
		return
	}
	updates := make(map[string]any, 6)
	if threadDelta != 0 {
		updates["thread_count"] = gorm.Expr("GREATEST(thread_count + ?, 0)", threadDelta)
		updates["week_thread_count"] = gorm.Expr("GREATEST(week_thread_count + ?, 0)", threadDelta)
	}
	if likeDelta != 0 {
		updates["like_count"] = gorm.Expr("GREATEST(like_count + ?, 0)", likeDelta)
		updates["week_like_count"] = gorm.Expr("GREATEST(week_like_count + ?, 0)", likeDelta)
	}
	if repostDelta != 0 {
		updates["repost_count"] = gorm.Expr("GREATEST(repost_count + ?, 0)", repostDelta)
		updates["week_repost_count"] = gorm.Expr("GREATEST(week_repost_count + ?, 0)", repostDelta)
	}
	db.Model(&threadmodel.Event{}).
		Where("id = ?", eventID).
		UpdateColumns(updates)
	InvalidateEventCounterCaches(c, eventID)
}

func PatchEventCounterByThreadID(db *gorm.DB, c *cache.Cache, threadID uint, column string, delta int) {
	if eventID, ok := EventIDForThread(db, threadID); ok {
		PatchEventCounter(db, c, eventID, column, delta)
		return
	}
	var thread threadmodel.Thread
	if err := db.Select("root_id", "parent_id", "original_thread_id").Where("thread_id = ?", threadID).First(&thread).Error; err != nil {
		return
	}
	if thread.RootID != nil {
		threadID = *thread.RootID
	} else if thread.ParentID != nil {
		threadID = *thread.ParentID
	} else if thread.OriginalThreadID != nil {
		threadID = *thread.OriginalThreadID
	} else {
		return
	}
	if eventID, ok := EventIDForThread(db, threadID); ok {
		PatchEventCounter(db, c, eventID, column, delta)
	}
}

func RefreshEventParticipantCount(db *gorm.DB, c *cache.Cache, eventID string) {
	if eventID == "" {
		return
	}
	db.Exec(`
		UPDATE events
		SET participant_count = (
			SELECT COUNT(DISTINCT threads.user_id)
			FROM thread_events
			JOIN threads ON threads.thread_id = thread_events.thread_id
			WHERE thread_events.event_id = ? AND threads.type <> ?
		)
		WHERE id = ?
	`, eventID, threadmodel.ThreadTypeRepost, eventID)
	InvalidateEventCounterCaches(c, eventID)
}

func InvalidateEventCounterCaches(c *cache.Cache, eventID string) {
	if c == nil {
		return
	}
	ctx := context.Background()
	_ = c.Delete(ctx, EventIDKey(eventID))
	_ = c.InvalidateTags(ctx, EventListsTag(), EventThreadsByEventTag(eventID))
}

func EventIDKey(id string) string {
	return fmt.Sprintf("event:id:%s", id)
}
