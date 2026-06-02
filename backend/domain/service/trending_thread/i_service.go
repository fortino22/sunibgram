package trendingthreadsvciface

import (
	"context"

	trendingthreadrepoiface "github.com/sunibgram/domain/repository/trending_thread"
)

type ITrendingThreadService interface {
	GetTopWeekly(ctx context.Context, viewerID uint) (trendingthreadrepoiface.TrendingThreadsResult, error)
}
