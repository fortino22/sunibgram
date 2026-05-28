package trendinguseruc

import (
	"context"
	"errors"
	"time"

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

func (s *TrendingUserService) GetMyHistory(ctx context.Context, viewerID uint, startDate, endDate *time.Time) (trendingusersvciface.MyHistoryResult, error) {
	months, err := s.snapshotRepo.FindAvailableMonths(ctx, viewerID)
	if err != nil {
		return trendingusersvciface.MyHistoryResult{}, err
	}

	var start, end time.Time
	if startDate != nil && endDate != nil {
		start = *startDate
		end = *endDate
	} else if len(months) > 0 {
		// Default: latest month in available data.
		latest := months[0]
		start = time.Date(latest.Year, time.Month(latest.Month), 1, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 1, -1)
	} else {
		return trendingusersvciface.MyHistoryResult{AvailableMonths: months}, nil
	}

	if end.Before(start) {
		return trendingusersvciface.MyHistoryResult{}, errors.New("end_date must be on or after start_date")
	}

	rows, err := s.snapshotRepo.FindByUserAndRange(ctx, viewerID, start, end)
	if err != nil {
		return trendingusersvciface.MyHistoryResult{}, err
	}

	return trendingusersvciface.MyHistoryResult{
		Rows:            rows,
		StartDate:       start,
		EndDate:         end,
		AvailableMonths: months,
	}, nil
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
