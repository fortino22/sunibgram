package repository

import (
	"fmt"
	"time"

	"github.com/sunibgram/domain/model"
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

const threadFeedTTL = 5 * time.Minute

func preloadThreadUser(db *gorm.DB) *gorm.DB {
	return db.Select("user_id", "name", "user_name", "is_premium")
}

// userJoinSelect returns a fresh sub-statement restricting the JOIN'd users
// table to safe public columns. Used as the conds argument of Joins("User", ...)
// so the LEFT JOIN selects only those columns (password_hash and email stay in
// the DB, never loaded into Go memory). Mirrors preloadThreadUser column set.
func userJoinSelect(db *gorm.DB) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).
		Model(&model.User{}).
		Select("user_id", "name", "user_name", "is_premium")
}

// Phase 2 — flat JOIN for the User relation, Preload for Community and the
// OriginalThread chain. The main thread's User is LEFT JOIN'd directly into
// the primary SELECT (eliminates one Preload round trip). OriginalThread is
// still preloaded because the visibility filter and self-referential alias
// make a flat JOIN risky; instead, the OriginalThread Preload itself uses
// Joins("User") inside the callback so the OriginalThread.User load is
// flattened into the OriginalThread query.
//
// Deep User.UserFaculty.Faculty chains intentionally NOT preloaded — Faculty
// + is_premium overlayed via user:meta:<uid> in FeedBuilder. Existing
// fallback in FeedBuilder.AuthorField (u.UserFaculty != nil) is dead in the
// happy path with these Preloads dropped.
//
// Community is kept as Preload to avoid a double JOIN with the raw
// visibility filter Joins("LEFT JOIN communities ...") that callers add via
// applyThreadVisibility.
func preloadThreadDetails(db *gorm.DB) *gorm.DB {
	return db.
		Joins("User", userJoinSelect(db)).
		Preload("Community").
		Preload("OriginalThread", func(otdb *gorm.DB) *gorm.DB {
			return otdb.Joins("User", userJoinSelect(otdb))
		}).
		Preload("OriginalThread.Community").
		Preload("ThreadEvents.Event").
		Preload("OriginalThread.ThreadEvents.Event")
}

func preloadThreadParentDetails(db *gorm.DB) *gorm.DB {
	return db.
		Joins("User", userJoinSelect(db)).
		Preload("Community").
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
