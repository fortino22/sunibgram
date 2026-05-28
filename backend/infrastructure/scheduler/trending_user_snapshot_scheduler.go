package scheduler

import (
	"context"
	"log"
	"time"

	trendingusersnapshotmodel "github.com/sunibgram/domain/model/trending_user_snapshot"
	trendinguserrepoiface "github.com/sunibgram/domain/repository/trending_user"
	trendingusersnapshotrepoiface "github.com/sunibgram/domain/repository/trending_user_snapshot"
)

type TrendingUserSnapshotScheduler struct {
	trendingRepo trendinguserrepoiface.ITrendingUserRepository
	snapshotRepo trendingusersnapshotrepoiface.ITrendingUserSnapshotRepository
}

func NewTrendingUserSnapshotScheduler(
	trendingRepo trendinguserrepoiface.ITrendingUserRepository,
	snapshotRepo trendingusersnapshotrepoiface.ITrendingUserSnapshotRepository,
) *TrendingUserSnapshotScheduler {
	return &TrendingUserSnapshotScheduler{trendingRepo: trendingRepo, snapshotRepo: snapshotRepo}
}

func (s *TrendingUserSnapshotScheduler) Start(ctx context.Context) {
	go s.loop(ctx)
}

func (s *TrendingUserSnapshotScheduler) loop(ctx context.Context) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.UTC
	}
	for {
		next := nextMondayAt(time.Now().In(loc), 0, 5, loc)
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Until(next)):
		}
		s.run(ctx, loc)
	}
}

// nextMondayAt returns the next instant at the given hour/minute on Monday in loc.
// Always strictly in the future.
func nextMondayAt(now time.Time, hour, minute int, loc *time.Location) time.Time {
	daysUntilMonday := (int(time.Monday) - int(now.Weekday()) + 7) % 7
	candidate := time.Date(now.Year(), now.Month(), now.Day()+daysUntilMonday, hour, minute, 0, 0, loc)
	if !candidate.After(now) {
		candidate = candidate.AddDate(0, 0, 7)
	}
	return candidate
}

func (s *TrendingUserSnapshotScheduler) run(ctx context.Context, loc *time.Location) {
	weekStart := time.Now().In(loc).AddDate(0, 0, -7)
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, loc)

	scores, err := s.trendingRepo.ComputeAllWeeklyScores(ctx)
	if err != nil {
		log.Printf("trending snapshot: compute scores failed: %v", err)
		return
	}
	if len(scores) == 0 {
		log.Printf("trending snapshot: no active users for week starting %s", weekStart.Format("2006-01-02"))
		return
	}

	rows := make([]trendingusersnapshotmodel.TrendingUserSnapshot, len(scores))
	for i, s := range scores {
		rows[i] = trendingusersnapshotmodel.TrendingUserSnapshot{
			UserID:       s.UserID,
			Rank:         i + 1,
			Score:        s.Score,
			LikesIn:      s.LikesIn,
			CommentsIn:   s.CommentsIn,
			BookmarksIn:  s.BookmarksIn,
			RepostsIn:    s.RepostsIn,
			ThreadCount:  s.ThreadCount,
			LikesOut:     s.LikesOut,
			CommentsOut:  s.CommentsOut,
			BookmarksOut: s.BookmarksOut,
			PostsOut:     s.PostsOut,
		}
	}
	if err := s.snapshotRepo.UpsertWeek(ctx, weekStart, rows); err != nil {
		log.Printf("trending snapshot: upsert week %s failed: %v", weekStart.Format("2006-01-02"), err)
		return
	}
	log.Printf("trending snapshot: persisted %d rows for week starting %s", len(rows), weekStart.Format("2006-01-02"))
}
