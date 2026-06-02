package trendingthreadrepoiface

import (
	"context"
	"time"

	trendingthreadmodel "github.com/sunibgram/domain/model/trending_thread"
)

type TrendingThreadsResult struct {
	Threads   []trendingthreadmodel.TrendingThread
	CachedAt  time.Time
	RefreshAt time.Time
}

type ITrendingThreadRepository interface {
	GetTopWeekly(ctx context.Context, viewerID uint, limit int) (TrendingThreadsResult, error)
}
