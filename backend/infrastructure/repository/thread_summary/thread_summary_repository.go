package threadsummaryrepo

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	threadsummarymodel "github.com/sunibgram/domain/model/thread_summary"
	threadsummaryrepoiface "github.com/sunibgram/domain/repository/thread_summary"
)

type ThreadSummaryRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewThreadSummaryRepository(writeDB, readDB *gorm.DB) threadsummaryrepoiface.IThreadSummaryRepository {
	return &ThreadSummaryRepository{writeDB: writeDB, readDB: readDB}
}

func (r *ThreadSummaryRepository) GetSummary(threadID uint) (*threadsummarymodel.ThreadSummary, error) {
	var summary threadsummarymodel.ThreadSummary
	err := r.readDB.Where("thread_id = ?", threadID).First(&summary).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &summary, nil
}

func (r *ThreadSummaryRepository) UpsertSummary(summary, contentSnapshot string, threadID uint) error {
	row := threadsummarymodel.ThreadSummary{
		ThreadID:        threadID,
		Summary:         summary,
		ContentSnapshot: contentSnapshot,
		Version:         1,
		IsStale:         false,
	}
	return r.writeDB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "thread_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"summary":          summary,
			"content_snapshot": contentSnapshot,
			"is_stale":         false,
			"version":          gorm.Expr("thread_summaries.version + 1"),
			"updated_at":       time.Now(),
		}),
	}).Create(&row).Error
}

func (r *ThreadSummaryRepository) MarkStale(threadID uint) error {
	return r.writeDB.Model(&threadsummarymodel.ThreadSummary{}).
		Where("thread_id = ?", threadID).
		Update("is_stale", true).Error
}

func (r *ThreadSummaryRepository) RecordUserView(userID, threadID uint) error {
	row := threadsummarymodel.UserThreadSummary{UserID: userID, ThreadID: threadID}
	return r.writeDB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "thread_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"updated_at": time.Now(),
		}),
	}).Create(&row).Error
}

func (r *ThreadSummaryRepository) ListUserHistory(userID uint, offset, limit int, q string) ([]threadsummarymodel.ThreadSummary, []time.Time, int64, error) {
	// A non-empty query filters on the joined summary text. Both the count and
	// the row query must join thread_summaries and apply the same predicate so
	// total matches the returned set (drives the infinite-scroll stop condition).
	pattern := "%" + q + "%"

	countQ := r.readDB.Model(&threadsummarymodel.UserThreadSummary{}).
		Where("user_id = ?", userID)
	if q != "" {
		countQ = countQ.
			Joins("JOIN thread_summaries AS ts ON ts.thread_id = user_thread_summaries.thread_id").
			Where("ts.content_snapshot ILIKE ? OR ts.summary ILIKE ?", pattern, pattern)
	}
	var total int64
	if err := countQ.Count(&total).Error; err != nil {
		return nil, nil, 0, err
	}
	if total == 0 {
		return []threadsummarymodel.ThreadSummary{}, []time.Time{}, 0, nil
	}

	// Join history rows to canonical summaries, newest-summarized first. Select
	// the summary columns plus the user's view timestamp in one pass.
	type row struct {
		threadsummarymodel.ThreadSummary
		SummarizedAt time.Time
	}
	rowQ := r.readDB.
		Table("user_thread_summaries AS uts").
		Select("ts.*, uts.updated_at AS summarized_at").
		Joins("JOIN thread_summaries AS ts ON ts.thread_id = uts.thread_id").
		Where("uts.user_id = ?", userID)
	if q != "" {
		rowQ = rowQ.Where("ts.content_snapshot ILIKE ? OR ts.summary ILIKE ?", pattern, pattern)
	}
	var rows []row
	err := rowQ.
		Order("uts.updated_at DESC").
		Offset(offset).Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, nil, 0, err
	}

	summaries := make([]threadsummarymodel.ThreadSummary, len(rows))
	timestamps := make([]time.Time, len(rows))
	for i, rw := range rows {
		summaries[i] = rw.ThreadSummary
		timestamps[i] = rw.SummarizedAt
	}
	return summaries, timestamps, total, nil
}
