package repository

import (
	"fmt"
	"time"

	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
)

type threadCached struct {
	db        *threadDB
	cache     *cache.Cache
	ttl       time.Duration
	likes     *interactionRepo
	bookmarks *interactionRepo
	reposts   *interactionRepo
}

type threadDB struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

const threadFeedTTL = 60 * time.Second

func preloadThreadUser(db *gorm.DB) *gorm.DB {
	return db.Select("user_id", "name", "user_name", "is_premium")
}

func preloadThreadDetails(db *gorm.DB) *gorm.DB {
	return db.
		Preload("User", preloadThreadUser).
		Preload("OriginalThread.User", preloadThreadUser).
		Preload("Community").
		Preload("OriginalThread.Community").
		Preload("User.UserFaculty.Faculty").
		Preload("OriginalThread.User.UserFaculty.Faculty").
		Preload("ThreadEvents.Event").
		Preload("OriginalThread.ThreadEvents.Event")
}

func preloadThreadParentDetails(db *gorm.DB) *gorm.DB {
	return db.
		Preload("User", preloadThreadUser).
		Preload("Community").
		Preload("User.UserFaculty.Faculty").
		Preload("ThreadEvents.Event")
}

func applyThreadVisibility(db *gorm.DB, requesterID *uint) *gorm.DB {
	db = db.Joins("LEFT JOIN user_settings thread_visibility_settings ON thread_visibility_settings.user_id = threads.user_id")
	if requesterID == nil {
		return db.Where("COALESCE(thread_visibility_settings.is_private, false) = ?", false)
	}
	db = db.Joins("LEFT JOIN user_relationships thread_visibility_relationships ON thread_visibility_relationships.follower_id = ? AND thread_visibility_relationships.following_id = threads.user_id", *requesterID)
	return db.Where(
		"threads.user_id = ? OR COALESCE(thread_visibility_settings.is_private, false) = ? OR thread_visibility_relationships.following_id IS NOT NULL",
		*requesterID,
		false,
	)
}

func threadVisibilityCachePart(requesterID *uint) string {
	if requesterID == nil {
		return "guest"
	}
	return fmt.Sprintf("%d", *requesterID)
}

func NewThreadRepository(writeDB, readDB *gorm.DB, c *cache.Cache) (domainrepo.IThreadRepository, func(uint, string, int)) {
	tdb := &threadDB{writeDB: writeDB, readDB: readDB}
	r := &threadCached{
		db:        tdb,
		cache:     c,
		ttl:       cache.DefaultTTL,
		likes:     newInteractionRepo(writeDB, readDB, c, "likes", tdb),
		bookmarks: newInteractionRepo(writeDB, readDB, c, "bookmarks", tdb),
		reposts:   newInteractionRepo(writeDB, readDB, c, "reposts", tdb),
	}
	return r, r.patchCachedCount
}
