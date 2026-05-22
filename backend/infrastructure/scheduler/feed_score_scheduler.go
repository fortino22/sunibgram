package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
)

type FeedScoreScheduler struct {
	db    *gorm.DB
	cache *cache.Cache
}

func NewFeedScoreScheduler(db *gorm.DB, c *cache.Cache) *FeedScoreScheduler {
	return &FeedScoreScheduler{db: db, cache: c}
}

func (s *FeedScoreScheduler) Start() {
	go func() {
		s.run()
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			s.run()
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

	if err := s.db.Exec("UPDATE threads SET feed_score = 0 WHERE created_at <= NOW() - INTERVAL '7 days'").Error; err != nil {
		log.Printf("feed score reset failed: %v", err)
		return
	}

	var threads []model.Thread
	if err := preloadSchedulerThreadDetails(s.db).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN user_settings ON user_settings.user_id = threads.user_id").
		Where("threads.created_at > NOW() - INTERVAL '7 days'").
		Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeReply)}).
		Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
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

	if err := s.cache.DeleteByPattern(context.Background(), "feed:foryou:*"); err != nil {
		log.Printf("feed cache invalidation failed: %v", err)
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
		Model(&model.User{}).
		Select("user_id", "name", "user_name", "is_premium")
}

// Phase 2 — User flattened into LEFT JOIN, OriginalThread kept as Preload
// with nested User Joins. Deep User.UserFaculty.Faculty intentionally not
// loaded — overlay via user:meta:<uid> in FeedBuilder handles it.
func preloadSchedulerThreadDetails(db *gorm.DB) *gorm.DB {
	return db.
		Joins("User", schedulerUserJoinSelect(db)).
		Preload("Community").
		Preload("OriginalThread", func(otdb *gorm.DB) *gorm.DB {
			return otdb.Joins("User", schedulerUserJoinSelect(otdb))
		}).
		Preload("OriginalThread.Community").
		Preload("ThreadEvents.Event").
		Preload("OriginalThread.ThreadEvents.Event")
}
