package trendreportuc

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	trendreportmodel "github.com/sunibgram/domain/model/trend_report"
	trendreportrepoiface "github.com/sunibgram/domain/repository/trend_report"
	trendreportsvciface "github.com/sunibgram/domain/service/trend_report"
	hashtagsvciface "github.com/sunibgram/domain/service/hashtag"
)

type trendReportService struct {
	repo       trendreportrepoiface.ITrendReportRepository
	hashtagSvc hashtagsvciface.IHashtagService
}

func NewTrendReportService(repo trendreportrepoiface.ITrendReportRepository, hashtagSvc hashtagsvciface.IHashtagService) trendreportsvciface.ITrendReportService {
	return &trendReportService{repo: repo, hashtagSvc: hashtagSvc}
}

func (s *trendReportService) CreateReport(ctx context.Context, reporterID uint, hashtagName string, reasonCode string) error {
	name := strings.TrimPrefix(strings.TrimSpace(hashtagName), "#")
	if name == "" {
		return errors.New("hashtag name is required")
	}
	code := strings.TrimSpace(reasonCode)
	if code == "" {
		return errors.New("reason code is required")
	}

	hashtagID, err := s.repo.FindHashtagIDByName(ctx, name)
	if err != nil {
		return err
	}
	if hashtagID == uuid.Nil {
		return errors.New("hashtag not found")
	}

	reason, err := s.repo.FindReasonByCode(ctx, code)
	if err != nil {
		return err
	}
	if reason == nil {
		return errors.New("reason not found")
	}

	if err := s.repo.Save(ctx, &trendreportmodel.TrendReport{
		ReporterID: reporterID,
		HashtagID:  hashtagID,
		ReasonID:   reason.ID,
	}); err != nil {
		return err
	}

	if s.hashtagSvc != nil {
		s.hashtagSvc.InvalidateTrendingCache(ctx)
	}
	return nil
}

func (s *trendReportService) ListReasons(ctx context.Context) ([]trendreportmodel.TrendReportReason, error) {
	return s.repo.ListReasons(ctx)
}

func (s *trendReportService) CountByReason(ctx context.Context) ([]trendreportrepoiface.ReasonCount, error) {
	return s.repo.CountByReason(ctx)
}
