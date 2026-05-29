package trendreportrepo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	hashtagmodel "github.com/sunibgram/domain/model/hashtag"
	trendreportmodel "github.com/sunibgram/domain/model/trend_report"
	trendreportrepoiface "github.com/sunibgram/domain/repository/trend_report"
)

type trendReportRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewTrendReportRepository(writeDB, readDB *gorm.DB) trendreportrepoiface.ITrendReportRepository {
	return &trendReportRepository{writeDB: writeDB, readDB: readDB}
}

func (r *trendReportRepository) Save(ctx context.Context, report *trendreportmodel.TrendReport) error {
	return r.writeDB.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(report).Error
}

func (r *trendReportRepository) FindReasonByCode(ctx context.Context, code string) (*trendreportmodel.TrendReportReason, error) {
	var reason trendreportmodel.TrendReportReason
	if err := r.readDB.WithContext(ctx).Where("code = ?", code).First(&reason).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &reason, nil
}

func (r *trendReportRepository) FindReasonByID(ctx context.Context, id uuid.UUID) (*trendreportmodel.TrendReportReason, error) {
	var reason trendreportmodel.TrendReportReason
	if err := r.readDB.WithContext(ctx).First(&reason, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &reason, nil
}

func (r *trendReportRepository) ListReasons(ctx context.Context) ([]trendreportmodel.TrendReportReason, error) {
	var reasons []trendreportmodel.TrendReportReason
	if err := r.readDB.WithContext(ctx).Order("label ASC").Find(&reasons).Error; err != nil {
		return nil, err
	}
	return reasons, nil
}

func (r *trendReportRepository) CountByReason(ctx context.Context) ([]trendreportrepoiface.ReasonCount, error) {
	var rows []trendreportrepoiface.ReasonCount
	err := r.readDB.WithContext(ctx).
		Table("trend_report_reasons AS r").
		Select("r.code AS code, r.label AS label, COUNT(tr.id) AS count").
		Joins("LEFT JOIN trend_reports tr ON tr.reason_id = r.id").
		Group("r.id, r.code, r.label").
		Order("count DESC, r.label ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *trendReportRepository) FindHashtagIDByName(ctx context.Context, name string) (uuid.UUID, error) {
	var hashtag hashtagmodel.Hashtag
	if err := r.readDB.WithContext(ctx).Select("id").Where("name = ?", name).First(&hashtag).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, nil
		}
		return uuid.Nil, err
	}
	return hashtag.ID, nil
}
