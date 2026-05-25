package scheduler

import (
	"context"
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"

	communitymodel "github.com/sunibgram/domain/model/community"
	threadmodel "github.com/sunibgram/domain/model/thread"
	"github.com/sunibgram/infrastructure/cache"
)

type FeedScoreScheduler struct {
	db    *gorm.DB
	cache *cache.Cache
}

func NewFeedScoreScheduler(db *gorm.DB, c *cache.Cache) *FeedScoreScheduler {
	return &FeedScoreScheduler{db: db, cache: c}
}

func (s *FeedScoreScheduler) Start(ctx context.Context) {
	go func() {
		s.run()
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.run()
			}
		}
	}()
}

func (s *FeedScoreScheduler) run() {
	// feed_score stores RAW weighted engagement (no decay). Read-time decay
	// (1 / (1 + age_h / forYouHalflifeHours)) is the only decay applied — see
	// FindForYouFeed.addCandidates. Storing the raw signal lets every feed
	// variant (For You, trending, search) apply its own decay curve without
	// re-running the scheduler, and keeps the stored value interpretable.
	//
	// IS DISTINCT FROM skips WAL writes for rows whose raw score did not move
	// since the last tick — most cold threads stay constant after their first
	// minute of life.
	if err := s.db.Exec(`
		UPDATE threads
		SET feed_score = new_score
		FROM (
			SELECT thread_id,
				(like_count::float + 3.0 * comment_count::float + 5.0 * repost_count::float + 4.0 * bookmark_count::float) AS new_score
			FROM threads
			WHERE created_at > NOW() - INTERVAL '7 days'
		) computed
		WHERE threads.thread_id = computed.thread_id
		  AND threads.feed_score IS DISTINCT FROM computed.new_score
	`).Error; err != nil {
		log.Printf("feed score recompute failed: %v", err)
		return
	}

	// Global trending now caches as a Redis ZSET of (thread_id → feed_score)
	// instead of a JSON-serialized []Thread. Two big wins:
	//   1. Redis payload: 100 thread IDs+scores ≈ 3 KB vs 100 fully-preloaded
	//      threads ≈ 200 KB (50–70× smaller).
	//   2. Per-request deserialization: readers only decode ints/floats; the
	//      full thread graph is hydrated via the existing per-thread cache
	//      (thread:id:<id>) on demand for the rows that actually surface.
	//
	// The thread:id:<id> cache is populated as a side effect of hydrateThreads
	// for any rows fetched from the DB, so steady-state reads serve trending
	// entirely from per-thread cache without re-querying threads.
	// Explicit gorm column tags so the Find() scan resolves correctly even if
	// GORM's NamingStrategy changes upstream. Fields project the only two
	// columns the ZSET cache needs.
	type trendingRow struct {
		ThreadID  uint    `gorm:"column:thread_id"`
		FeedScore float64 `gorm:"column:feed_score"`
	}
	var rows []trendingRow
	if err := s.db.Model(&threadmodel.Thread{}).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN user_settings ON user_settings.user_id = threads.user_id").
		Where("threads.created_at > NOW() - INTERVAL '7 days'").
		Where("threads.type IN ?", []string{string(threadmodel.ThreadTypePost), string(threadmodel.ThreadTypeReply)}).
		Where("threads.community_id IS NULL OR communities.visibility = ?", communitymodel.CommunityVisibilityPublic).
		Where("COALESCE(user_settings.is_private, false) = ?", false).
		Where("threads.parent_id IS NULL OR (threads.like_count + threads.repost_count + threads.comment_count) > 0").
		Select("threads.thread_id, threads.feed_score").
		Order("threads.feed_score DESC").
		Limit(100).
		Find(&rows).Error; err != nil {
		log.Printf("global trending feed refresh failed: %v", err)
		return
	}

	members := make([]cache.ZSetMember, len(rows))
	for i, row := range rows {
		members[i] = cache.ZSetMember{
			Member: strconv.FormatUint(uint64(row.ThreadID), 10),
			Score:  row.FeedScore,
		}
	}
	if err := s.cache.ZAddReplace(context.Background(), "feed:trending:global", members, 5*time.Minute); err != nil {
		log.Printf("global trending feed cache set failed: %v", err)
		return
	}

	log.Println("feed scores recomputed")
}

