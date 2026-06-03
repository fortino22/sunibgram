package foxhistoryrepo

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	foxhistorymodel "github.com/sunibgram/domain/model/fox_history"
	foxhistoryrepoiface "github.com/sunibgram/domain/repository/fox_history"
)

type FoxHistoryRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewFoxHistoryRepository(writeDB, readDB *gorm.DB) foxhistoryrepoiface.IFoxHistoryRepository {
	return &FoxHistoryRepository{writeDB: writeDB, readDB: readDB}
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
	return r.writeDB.Clauses(clause.OnConflict{
		Columns:      []clause.Column{{Name: "user_id"}, {Name: "thread_id"}},
		TargetWhere:  clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "type", Value: foxhistorymodel.TypeSummary}}},
		DoUpdates: clause.Assignments(map[string]any{
			"body":       summary,
			"snapshot":   snapshot,
			"updated_at": time.Now(),
		}),
	}).Create(&row).Error
}

func (r *FoxHistoryRepository) RecordChat(userID uint, question, answer string) error {
	row := foxhistorymodel.FoxHistory{
		UserID:   userID,
		Type:     foxhistorymodel.TypeChat,
		Question: question,
		Body:     answer,
	}
	return r.writeDB.Create(&row).Error
}

func (r *FoxHistoryRepository) ListHistory(userID uint, offset, limit int, q, typeFilter string) ([]foxhistorymodel.FoxHistory, int64, error) {
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
	if total == 0 {
		return []foxhistorymodel.FoxHistory{}, 0, nil
	}

	var rows []foxhistorymodel.FoxHistory
	err := base.Order("updated_at DESC").Offset(offset).Limit(limit).Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
