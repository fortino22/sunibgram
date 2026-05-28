package trendreportsvciface

import (
	"context"

	trendreportmodel "github.com/sunibgram/domain/model/trend_report"
	trendreportrepoiface "github.com/sunibgram/domain/repository/trend_report"
)

type ITrendReportService interface {
	CreateReport(ctx context.Context, reporterID uint, hashtagName string, reasonCode string) error
	ListReasons(ctx context.Context) ([]trendreportmodel.TrendReportReason, error)
	CountByReason(ctx context.Context) ([]trendreportrepoiface.ReasonCount, error)
}
