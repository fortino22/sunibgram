package trendreportrepoiface

import (
	"context"

	"github.com/google/uuid"

	trendreportmodel "github.com/sunibgram/domain/model/trend_report"
)

type ReasonCount struct {
	Code  string
	Label string
	Count int64
}

type ITrendReportRepository interface {
	Save(ctx context.Context, report *trendreportmodel.TrendReport) error
	FindReasonByCode(ctx context.Context, code string) (*trendreportmodel.TrendReportReason, error)
	FindReasonByID(ctx context.Context, id uuid.UUID) (*trendreportmodel.TrendReportReason, error)
	ListReasons(ctx context.Context) ([]trendreportmodel.TrendReportReason, error)
	CountByReason(ctx context.Context) ([]ReasonCount, error)
	FindHashtagIDByName(ctx context.Context, name string) (uuid.UUID, error)
}
