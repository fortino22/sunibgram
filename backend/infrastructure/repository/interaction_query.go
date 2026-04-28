package repository

import (
	"time"

	"github.com/sunibgram/domain/model"
	"gorm.io/gorm"
)

type interactionTimestamp struct {
	ThreadID      uint
	InteractedAt  time.Time
}

func findThreadsByInteraction(db *gorm.DB, table string, userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error) {
	var total int64
	db.Table(table).Where("user_id = ?", userID).Count(&total)

	var rows []interactionTimestamp
	offset := (page - 1) * limit
	err := db.Table(table).
		Select("thread_id, created_at AS interacted_at").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return nil, nil, total, err
	}

	ids := make([]uint, len(rows))
	orderMap := make(map[uint]time.Time, len(rows))
	for i, r := range rows {
		ids[i] = r.ThreadID
		orderMap[r.ThreadID] = r.InteractedAt
	}

	var threads []model.Thread
	if err := db.Preload("User").Where("thread_id IN ?", ids).Find(&threads).Error; err != nil {
		return nil, nil, total, err
	}

	threadMap := make(map[uint]model.Thread, len(threads))
	for _, t := range threads {
		threadMap[t.ThreadID] = t
	}
	sorted := make([]model.Thread, len(ids))
	timestamps := make([]time.Time, len(ids))
	for i, id := range ids {
		sorted[i] = threadMap[id]
		timestamps[i] = orderMap[id]
	}

	return sorted, timestamps, total, nil
}
