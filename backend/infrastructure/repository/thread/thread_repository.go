package threadrepo

import (
	"fmt"
	"time"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"

	usermodel "github.com/sunibgram/domain/model/user"
	threadrepoiface "github.com/sunibgram/domain/repository/thread"
	"github.com/sunibgram/infrastructure/cache"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
)

type threadCached struct {
	db        *ThreadDB
	cache     *cache.Cache
	ttl       time.Duration
	likes     *infrarepoShared.InteractionRepo
	bookmarks *infrarepoShared.InteractionRepo
	reposts   *infrarepoShared.InteractionRepo
	// forYouSingleflight coalesces concurrent cache-miss requests against the
	// same For You source key onto a single underlying DB query. Eliminates
	// thundering-herd at TTL expiry.
	forYouSingleflight singleflight.Group
	// forYouMetrics records cache hit/miss + DB latency counters for the For
	// You pipeline. Read via the exposed Snapshot() accessor.
	forYouMetrics forYouMetrics
}

// ForYouMetricsSnapshot returns a copy of the For You feed cache/latency
// counters. Intended for an admin observability endpoint; safe to call
// concurrently with feed traffic.
func (r *threadCached) ForYouMetricsSnapshot() ForYouMetricsSnapshot {
	return r.forYouMetrics.Snapshot()
}

type ThreadDB struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

const ThreadFeedTTL = 5 * time.Minute

func preloadThreadUser(db *gorm.DB) *gorm.DB {
	return db.Select("user_id", "name", "user_name", "is_premium")
}

// userJoinSelect returns a fresh sub-statement restricting the JOIN'd users
// table to safe public columns. Used as the conds argument of Joins("User", ...)
// so the LEFT JOIN selects only those columns (password_hash and email stay in
// the DB, never loaded into Go memory). Mirrors preloadThreadUser column set.
func userJoinSelect(db *gorm.DB) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).
		Model(&usermodel.User{}).
		Select("user_id", "name", "user_name", "is_premium")
}

// Phase 3 — single projected JOIN for all 1:1 relations. Main thread row +
// User + Community + OriginalThread + OriginalThread.User +
// OriginalThread.Community land in ONE SQL query via GORM relation joins.
// Only 1:N ThreadEvents.Event chains stay as Preload (cannot flatten without
// row explosion).
//
// Join ordering matters: parent relations must be registered BEFORE child
// relations that share a path. Joins("OriginalThread") is listed before
// Joins("OriginalThread.User", ...) so GORM's iteration skips re-adding the
// OriginalThread level and applies the User-column Selects only to the User
// leaf (see gorm/callbacks/query.go:242 specifiedRelationsName check).
//
// userJoinSelect restricts the User column SELECT to (user_id, name,
// user_name, is_premium) — same projection as the legacy preloadThreadUser.
// Password / email stay in the DB, never loaded into Go memory.
//
// Deep User.UserFaculty.Faculty chains intentionally NOT joined — faculty +
// is_premium are overlayed via user:meta:<uid> in FeedBuilder.
//
// Note: callers still add a raw `Joins("LEFT JOIN communities ON ...")` via
// ApplyThreadVisibility for the unaliased visibility filter. That coexists
// with the aliased "Community" relation join here (two LEFT JOINs to
// communities, different aliases). Postgres handles this trivially.
func PreloadThreadDetails(db *gorm.DB) *gorm.DB {
	return db.
		Joins("User", userJoinSelect(db)).
		Joins("Community").
		Joins("OriginalThread").
		Joins("OriginalThread.User", userJoinSelect(db)).
		Joins("OriginalThread.Community").
		Preload("ThreadEvents.Event").
		Preload("OriginalThread.ThreadEvents.Event")
}

func preloadThreadParentDetails(db *gorm.DB) *gorm.DB {
	return db.
		Joins("User", userJoinSelect(db)).
		Joins("Community").
		Preload("ThreadEvents.Event")
}

func ApplyThreadVisibility(db *gorm.DB, requesterID *uint) *gorm.DB {
	db = db.Joins("LEFT JOIN user_settings thread_visibility_settings ON thread_visibility_settings.user_id = threads.user_id")
	if requesterID == nil {
		return db.Where("COALESCE(thread_visibility_settings.is_private, false) = ?", false)
	}
	db = ApplyBlockFilter(db, *requesterID)
	db = db.Joins("LEFT JOIN user_relationships thread_visibility_relationships ON thread_visibility_relationships.follower_id = ? AND thread_visibility_relationships.following_id = threads.user_id", *requesterID)
	return db.Where(
		"threads.user_id = ? OR COALESCE(thread_visibility_settings.is_private, false) = ? OR thread_visibility_relationships.following_id IS NOT NULL",
		*requesterID,
		false,
	)
}

// ApplyBlockFilter excludes threads authored by users blocked by the requester
// or who have blocked the requester (mutual hide). Use it directly on feed
// queries that build their own visibility joins instead of ApplyThreadVisibility.
func ApplyBlockFilter(db *gorm.DB, requesterID uint) *gorm.DB {
	return ApplyBlockFilterOnColumn(db, "threads.user_id", requesterID)
}

// ApplyBlockFilterOnColumn is ApplyBlockFilter for queries whose author/user id
// lives in a different column (e.g. users.user_id in explore searches). It also
// excludes authors the requester has muted (one-way), so every thread-visibility
// consumer hides muted users' posts alongside blocked users'.
func ApplyBlockFilterOnColumn(db *gorm.DB, userColumn string, requesterID uint) *gorm.DB {
	return db.Where(userColumn+` NOT IN (
		SELECT blocked_id FROM user_blocks WHERE blocker_id = ?
		UNION
		SELECT blocker_id FROM user_blocks WHERE blocked_id = ?)`, requesterID, requesterID).
		Where(userColumn+` NOT IN (
		SELECT muted_user_id FROM notification_user_mutes WHERE user_id = ?)`, requesterID)
}

func ThreadVisibilityCachePart(requesterID *uint) string {
	if requesterID == nil {
		return "guest"
	}
	return fmt.Sprintf("%d", *requesterID)
}

func NewThreadRepository(writeDB, readDB *gorm.DB, c *cache.Cache) (threadrepoiface.IThreadRepository, func(uint, string, int)) {
	tdb := &ThreadDB{writeDB: writeDB, readDB: readDB}
	r := &threadCached{
		db:        tdb,
		cache:     c,
		ttl:       cache.DefaultTTL,
		likes:     infrarepoShared.NewInteractionRepo(writeDB, readDB, c, "likes", PreloadThreadDetails),
		bookmarks: infrarepoShared.NewInteractionRepo(writeDB, readDB, c, "bookmarks", PreloadThreadDetails),
		reposts:   infrarepoShared.NewInteractionRepo(writeDB, readDB, c, "reposts", PreloadThreadDetails),
	}
	return r, r.patchCachedCount
}
