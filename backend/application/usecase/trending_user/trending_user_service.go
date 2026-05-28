package trendinguseruc

import (
	"context"

	"github.com/sunibgram/application/helper"
	trendinguserrepoiface "github.com/sunibgram/domain/repository/trending_user"
	trendingusersvciface "github.com/sunibgram/domain/service/trending_user"
)

const defaultTrendingUserLimit = 20

type TrendingUserService struct {
	repo trendinguserrepoiface.ITrendingUserRepository
}

func NewTrendingUserService(repo trendinguserrepoiface.ITrendingUserRepository) trendingusersvciface.ITrendingUserService {
	return &TrendingUserService{repo: repo}
}

func (s *TrendingUserService) GetWeeklyTrending(ctx context.Context, viewerID uint, page, limit int) (trendinguserrepoiface.TrendingUsersPage, error) {
	page, limit = helper.NormalizePageLimit(page, limit, defaultTrendingUserLimit)
	return s.repo.GetWeeklyTrending(ctx, viewerID, page, limit)
}
