package foxhistoryrepo

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	foxhistorymodel "github.com/sunibgram/domain/model/fox_history"
	foxhistoryrepoiface "github.com/sunibgram/domain/repository/fox_history"
	"github.com/sunibgram/infrastructure/cache"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
)

const foxHistoryTTL = 5 * time.Minute

type FoxHistoryRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewFoxHistoryRepository(writeDB, readDB *gorm.DB, c *cache.Cache) foxhistoryrepoiface.IFoxHistoryRepository {
	return &FoxHistoryRepository{writeDB: writeDB, readDB: readDB, cache: c}
}

func foxHistoryListKey(userID uint, offset, limit int, q, typeFilter string) string {
	return fmt.Sprintf("fox_history:user:%d:type:%s:q:%s:offset:%d:limit:%d", userID, typeFilter, q, offset, limit)
}

type listCachePayload struct {
	Items []foxhistorymodel.FoxHistory `json:"items"`
	Total int64                        `json:"total"`
}

func (r *FoxHistoryRepository) RecordSummary(userID, threadID uint, summary, snapshot string) error {
	tid := threadID
	row := foxhistorymodel.FoxHistory{
		UserID:   userID,
		Type:     foxhistorymodel.TypeSummary,
		ThreadID: &tid,
		Body:     summary,
		Snapshot: snapshot,
	}
	// Conflict target is the partial unique index on (user_id, thread_id) WHERE
	// type='summary' (created in migration). The matching IndexWhere lets Postgres
	// pick that partial index as the arbiter.
	if err := r.writeDB.Clauses(clause.OnConflict{
		Columns:     []clause.Column{{Name: "user_id"}, {Name: "thread_id"}},
		TargetWhere: clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "type", Value: foxhistorymodel.TypeSummary}}},
		DoUpdates: clause.Assignments(map[string]any{
			"body":       summary,
			"snapshot":   snapshot,
			"updated_at": time.Now(),
		}),
	}).Create(&row).Error; err != nil {
		return err
	}
	r.invalidateUser(userID)
	return nil
}

func (r *FoxHistoryRepository) RecordChat(userID uint, question, answer string) error {
	row := foxhistorymodel.FoxHistory{
		UserID:   userID,
		Type:     foxhistorymodel.TypeChat,
		Question: question,
		Body:     answer,
	}
	if err := r.writeDB.Create(&row).Error; err != nil {
		return err
	}
	r.invalidateUser(userID)
	return nil
}

func (r *FoxHistoryRepository) ListHistory(userID uint, offset, limit int, q, typeFilter string) ([]foxhistorymodel.FoxHistory, int64, error) {
	ctx := context.Background()
	key := foxHistoryListKey(userID, offset, limit, q, typeFilter)
	if r.cache != nil {
		var payload listCachePayload
		if err := r.cache.Get(ctx, key, &payload); err == nil {
			return payload.Items, payload.Total, nil
		}
	}

	base := r.readDB.Model(&foxhistorymodel.FoxHistory{}).Where("user_id = ?", userID)
	if typeFilter != "" {
		base = base.Where("type = ?", typeFilter)
	}
	if q != "" {
		pattern := "%" + q + "%"
		base = base.Where("question ILIKE ? OR body ILIKE ? OR snapshot ILIKE ?", pattern, pattern, pattern)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := []foxhistorymodel.FoxHistory{}
	if total > 0 {
		if err := base.Order("updated_at DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
			return nil, 0, err
		}
	}

	if r.cache != nil {
		_ = r.cache.Set(ctx, key, listCachePayload{Items: rows, Total: total}, foxHistoryTTL)
		_ = r.cache.RegisterCacheKey(ctx, infrarepoShared.FoxHistoryUserTag(userID), key)
	}
	return rows, total, nil
}

func (r *FoxHistoryRepository) invalidateUser(userID uint) {
	if r.cache == nil {
		return
	}
	_ = r.cache.InvalidateTag(context.Background(), infrarepoShared.FoxHistoryUserTag(userID))
}
