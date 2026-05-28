package trendinguseruc

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/sunibgram/application/helper"
	trendingusersnapshotmodel "github.com/sunibgram/domain/model/trending_user_snapshot"
	trendinguserrepoiface "github.com/sunibgram/domain/repository/trending_user"
	trendingusersnapshotrepoiface "github.com/sunibgram/domain/repository/trending_user_snapshot"
	trendingusersvciface "github.com/sunibgram/domain/service/trending_user"
)

const (
	defaultTrendingUserLimit = 20
	defaultHistoryLimit      = 30
)

type TrendingUserService struct {
	repo         trendinguserrepoiface.ITrendingUserRepository
	snapshotRepo trendingusersnapshotrepoiface.ITrendingUserSnapshotRepository
}

func NewTrendingUserService(
	repo trendinguserrepoiface.ITrendingUserRepository,
	snapshotRepo trendingusersnapshotrepoiface.ITrendingUserSnapshotRepository,
) trendingusersvciface.ITrendingUserService {
	return &TrendingUserService{repo: repo, snapshotRepo: snapshotRepo}
}

func (s *TrendingUserService) GetWeeklyTrending(ctx context.Context, viewerID uint, page, limit int) (trendinguserrepoiface.TrendingUsersPage, error) {
	page, limit = helper.NormalizePageLimit(page, limit, defaultTrendingUserLimit)
	return s.repo.GetWeeklyTrending(ctx, viewerID, page, limit)
}

func (s *TrendingUserService) GetMyCurrent(ctx context.Context, viewerID uint) (trendingusersvciface.MyCurrentResult, error) {
	user, rank, page, err := s.repo.GetMyWeeklyTrending(ctx, viewerID)
	if err != nil {
		return trendingusersvciface.MyCurrentResult{}, err
	}
	return trendingusersvciface.MyCurrentResult{User: user, Rank: rank, Page: page}, nil
}

func (s *TrendingUserService) GetMyHistory(ctx context.Context, viewerID uint, year, month int) (trendingusersvciface.MyHistoryResult, error) {
	months, err := s.snapshotRepo.FindAvailableMonths(ctx, viewerID)
	if err != nil {
		return trendingusersvciface.MyHistoryResult{}, err
	}

	if (year == 0 || month == 0) && len(months) > 0 {
		year = months[0].Year
		month = months[0].Month
	}

	var rows []trendingusersnapshotmodel.TrendingUserSnapshot
	if year > 0 && month > 0 {
		rows, err = s.snapshotRepo.FindByUserAndMonth(ctx, viewerID, year, month)
		if err != nil {
			return trendingusersvciface.MyHistoryResult{}, err
		}
	}

	return trendingusersvciface.MyHistoryResult{Rows: rows, AvailableMonths: months}, nil
}

func (s *TrendingUserService) GetSnapshotByID(ctx context.Context, viewerID, snapshotID uint) (*trendingusersnapshotmodel.TrendingUserSnapshot, error) {
	row, err := s.snapshotRepo.FindByID(ctx, snapshotID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("snapshot not found")
		}
		return nil, err
	}
	if row.UserID != viewerID {
		return nil, errors.New("forbidden")
	}
	return row, nil
}
