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
