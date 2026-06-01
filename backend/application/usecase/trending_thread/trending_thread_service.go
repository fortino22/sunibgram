package trendingthreaduc

import (
	"context"

	trendingthreadrepoiface "github.com/sunibgram/domain/repository/trending_thread"
	trendingthreadsvciface "github.com/sunibgram/domain/service/trending_thread"
)

// topTrendingThreadLimit caps the leaderboard at the top 3 threads of the week.
const topTrendingThreadLimit = 3

type TrendingThreadService struct {
	repo trendingthreadrepoiface.ITrendingThreadRepository
}

func NewTrendingThreadService(repo trendingthreadrepoiface.ITrendingThreadRepository) trendingthreadsvciface.ITrendingThreadService {
	return &TrendingThreadService{repo: repo}
}

func (s *TrendingThreadService) GetTopWeekly(ctx context.Context, viewerID uint) (trendingthreadrepoiface.TrendingThreadsResult, error) {
	return s.repo.GetTopWeekly(ctx, viewerID, topTrendingThreadLimit)
}
