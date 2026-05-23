package scheduler

import (
	"context"
	"log"
	"time"

	"gorm.io/gorm"

	communitymodel "github.com/sunibgram/domain/model/community"
	threadmodel "github.com/sunibgram/domain/model/thread"
	usermodel "github.com/sunibgram/domain/model/user"
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
	if err := s.db.Exec(`
		UPDATE threads
		SET feed_score = (
			like_count::float +
			3.0 * comment_count::float +
			5.0 * repost_count::float +
			4.0 * bookmark_count::float
		) / (
			1.0 + POWER(EXTRACT(EPOCH FROM (NOW() - created_at)) / 3600.0, 1.5)
		)
		WHERE created_at > NOW() - INTERVAL '7 days'
	`).Error; err != nil {
		log.Printf("feed score recompute failed: %v", err)
		return
	}

	if err := s.db.Exec("UPDATE threads SET feed_score = 0 WHERE created_at <= NOW() - INTERVAL '7 days' AND feed_score <> 0").Error; err != nil {
		log.Printf("feed score reset failed: %v", err)
		return
	}

	var threads []threadmodel.Thread
	if err := preloadSchedulerThreadDetails(s.db).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN user_settings ON user_settings.user_id = threads.user_id").
		Where("threads.created_at > NOW() - INTERVAL '7 days'").
		Where("threads.type IN ?", []string{string(threadmodel.ThreadTypePost), string(threadmodel.ThreadTypeReply)}).
		Where("threads.community_id IS NULL OR communities.visibility = ?", communitymodel.CommunityVisibilityPublic).
		Where("COALESCE(user_settings.is_private, false) = ?", false).
		Where("threads.parent_id IS NULL OR (threads.like_count + threads.repost_count + threads.comment_count) > 0").
		Order("threads.feed_score DESC").
		Limit(100).
		Find(&threads).Error; err != nil {
		log.Printf("global trending feed refresh failed: %v", err)
		return
	}

	if err := s.cache.Set(context.Background(), "feed:trending:global", threads, 5*time.Minute); err != nil {
		log.Printf("global trending feed cache set failed: %v", err)
		return
	}

	log.Println("feed scores recomputed")
}

func preloadSchedulerThreadUser(db *gorm.DB) *gorm.DB {
	return db.Select("user_id", "name", "user_name", "is_premium")
}

// schedulerUserJoinSelect mirrors userJoinSelect in the repository package
// for the scheduler's Joins("User", ...) form. Restricts loaded user columns
// to public fields.
func schedulerUserJoinSelect(db *gorm.DB) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).
		Model(&usermodel.User{}).
		Select("user_id", "name", "user_name", "is_premium")
}

// Phase 3 — all 1:1 relations flattened into one JOIN query. Mirrors the
// repository-package preloadThreadDetails. Only ThreadEvents.Event chains
// remain as Preload (1:N, can't flatten). Join ordering matters: parent
// relations registered before child paths so GORM skips re-adding parent
// when traversing the nested Joins("OriginalThread.User", ...) chain.
func preloadSchedulerThreadDetails(db *gorm.DB) *gorm.DB {
	return db.
		Joins("User", schedulerUserJoinSelect(db)).
		Joins("Community").
		Joins("OriginalThread").
		Joins("OriginalThread.User", schedulerUserJoinSelect(db)).
		Joins("OriginalThread.Community").
		Preload("ThreadEvents.Event").
		Preload("OriginalThread.ThreadEvents.Event")
}
