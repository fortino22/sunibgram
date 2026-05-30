package trendinguserrepo

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	trendingusermodel "github.com/sunibgram/domain/model/trending_user"
	trendinguserrepoiface "github.com/sunibgram/domain/repository/trending_user"
	"github.com/sunibgram/infrastructure/cache"
)

const trendingUserTTL = 5 * time.Minute

type trendingUserRepository struct {
	readDB *gorm.DB
	cache  *cache.Cache
}

// scoreRow holds the cached, viewer-agnostic aggregate per user. User-mutable
// fields (name, username, avatar, is_premium) and the per-viewer is_following
// flag are resolved fresh on each request.
type scoreRow struct {
	UserID       uint    `json:"u"`
	Score        float64 `json:"s"`
	ThreadCount  int     `json:"tc"`
	LikesIn      int     `json:"li"`
	CommentsIn   int     `json:"ci"`
	BookmarksIn  int     `json:"bi"`
	RepostsIn    int     `json:"ri"`
	LikesOut     int     `json:"lo"`
	CommentsOut  int     `json:"co"`
	BookmarksOut int     `json:"bo"`
	PostsOut     int     `json:"po"`
}

type cachedTrendingScores struct {
	Rows     []scoreRow `json:"rows"`
	Total    int64      `json:"total"`
	CachedAt time.Time  `json:"cached_at"`
}

type userInfoRow struct {
	UserID           uint
	Username         string
	Name             string
	IsPremium        bool
	AvatarObjectName string
}

func NewTrendingUserRepository(readDB *gorm.DB, c *cache.Cache) trendinguserrepoiface.ITrendingUserRepository {
	return &trendingUserRepository{readDB: readDB, cache: c}
}

func scoresCacheKey(page, limit int) string {
	return fmt.Sprintf("trending:users:weekly:scores:p%d:l%d", page, limit)
}

const trendingUserScoresSQL = `
WITH window_start AS (SELECT NOW() - INTERVAL '7 days' AS ts),
received AS (
    SELECT t.user_id,
        COUNT(DISTINCT CASE WHEN t.type = 'POST' THEN t.thread_id END) AS thread_count,
        COALESCE(SUM(l.cnt), 0) AS likes_in,
        COALESCE(SUM(b.cnt), 0) AS bookmarks_in,
        COALESCE(SUM(rp.cnt), 0) AS reposts_in,
        COALESCE(SUM(cm.cnt), 0) AS comments_in
    FROM threads t
    LEFT JOIN (
        SELECT thread_id, COUNT(*) AS cnt FROM likes
        WHERE created_at > (SELECT ts FROM window_start)
        GROUP BY thread_id
    ) l ON l.thread_id = t.thread_id
    LEFT JOIN (
        SELECT thread_id, COUNT(*) AS cnt FROM bookmarks
        WHERE created_at > (SELECT ts FROM window_start)
        GROUP BY thread_id
    ) b ON b.thread_id = t.thread_id
    LEFT JOIN (
        SELECT thread_id, COUNT(*) AS cnt FROM reposts
        WHERE created_at > (SELECT ts FROM window_start)
        GROUP BY thread_id
    ) rp ON rp.thread_id = t.thread_id
    LEFT JOIN (
        SELECT parent_id AS thread_id, COUNT(*) AS cnt FROM threads
        WHERE type = 'REPLY' AND parent_id IS NOT NULL
          AND created_at > (SELECT ts FROM window_start)
        GROUP BY parent_id
    ) cm ON cm.thread_id = t.thread_id
    WHERE t.created_at > (SELECT ts FROM window_start)
       OR l.cnt IS NOT NULL OR b.cnt IS NOT NULL
       OR rp.cnt IS NOT NULL OR cm.cnt IS NOT NULL
    GROUP BY t.user_id
),
given_acts AS (
    SELECT user_id,
        COUNT(*) FILTER (WHERE src = 'like')     AS likes_out,
        COUNT(*) FILTER (WHERE src = 'bookmark') AS bookmarks_out,
        COUNT(*) FILTER (WHERE src = 'comment')  AS comments_out,
        COUNT(*) FILTER (WHERE src = 'post')     AS posts_out
    FROM (
        SELECT user_id, 'like' AS src FROM likes
        WHERE created_at > (SELECT ts FROM window_start)
        UNION ALL
        SELECT user_id, 'bookmark' FROM bookmarks
        WHERE created_at > (SELECT ts FROM window_start)
        UNION ALL
        SELECT user_id, 'comment' FROM threads
        WHERE type = 'REPLY' AND created_at > (SELECT ts FROM window_start)
        UNION ALL
        SELECT user_id, 'post' FROM threads
        WHERE type = 'POST' AND created_at > (SELECT ts FROM window_start)
    ) u
    GROUP BY user_id
)
SELECT u.user_id,
    COALESCE(rc.thread_count, 0)  AS thread_count,
    COALESCE(rc.likes_in, 0)      AS likes_in,
    COALESCE(rc.comments_in, 0)   AS comments_in,
    COALESCE(rc.bookmarks_in, 0)  AS bookmarks_in,
    COALESCE(rc.reposts_in, 0)    AS reposts_in,
    COALESCE(gv.likes_out, 0)     AS likes_out,
    COALESCE(gv.comments_out, 0)  AS comments_out,
    COALESCE(gv.bookmarks_out, 0) AS bookmarks_out,
    COALESCE(gv.posts_out, 0)     AS posts_out,
    (
        COALESCE(rc.likes_in, 0)
        + 3 * COALESCE(rc.comments_in, 0)
        + 4 * COALESCE(rc.bookmarks_in, 0)
        + 5 * COALESCE(rc.reposts_in, 0)
        + 0.2 * (
            COALESCE(gv.likes_out, 0)
            + 3 * COALESCE(gv.comments_out, 0)
            + 4 * COALESCE(gv.bookmarks_out, 0)
            + COALESCE(gv.posts_out, 0)
        )
    ) AS score
FROM users u
LEFT JOIN received   rc ON rc.user_id = u.user_id
LEFT JOIN given_acts gv ON gv.user_id = u.user_id
WHERE (rc.user_id IS NOT NULL OR gv.user_id IS NOT NULL)
ORDER BY score DESC, u.user_id ASC
LIMIT ? OFFSET ?
`

const trendingUserAllScoresSQL = `
WITH window_start AS (SELECT NOW() - INTERVAL '7 days' AS ts),
received AS (
    SELECT t.user_id,
        COUNT(DISTINCT CASE WHEN t.type = 'POST' THEN t.thread_id END) AS thread_count,
        COALESCE(SUM(l.cnt), 0) AS likes_in,
        COALESCE(SUM(b.cnt), 0) AS bookmarks_in,
        COALESCE(SUM(rp.cnt), 0) AS reposts_in,
        COALESCE(SUM(cm.cnt), 0) AS comments_in
    FROM threads t
    LEFT JOIN (
        SELECT thread_id, COUNT(*) AS cnt FROM likes
        WHERE created_at > (SELECT ts FROM window_start)
        GROUP BY thread_id
    ) l ON l.thread_id = t.thread_id
    LEFT JOIN (
        SELECT thread_id, COUNT(*) AS cnt FROM bookmarks
        WHERE created_at > (SELECT ts FROM window_start)
        GROUP BY thread_id
    ) b ON b.thread_id = t.thread_id
    LEFT JOIN (
        SELECT thread_id, COUNT(*) AS cnt FROM reposts
        WHERE created_at > (SELECT ts FROM window_start)
        GROUP BY thread_id
    ) rp ON rp.thread_id = t.thread_id
    LEFT JOIN (
        SELECT parent_id AS thread_id, COUNT(*) AS cnt FROM threads
        WHERE type = 'REPLY' AND parent_id IS NOT NULL
          AND created_at > (SELECT ts FROM window_start)
        GROUP BY parent_id
    ) cm ON cm.thread_id = t.thread_id
    WHERE t.created_at > (SELECT ts FROM window_start)
       OR l.cnt IS NOT NULL OR b.cnt IS NOT NULL
       OR rp.cnt IS NOT NULL OR cm.cnt IS NOT NULL
    GROUP BY t.user_id
),
given_acts AS (
    SELECT user_id,
        COUNT(*) FILTER (WHERE src = 'like')     AS likes_out,
        COUNT(*) FILTER (WHERE src = 'bookmark') AS bookmarks_out,
        COUNT(*) FILTER (WHERE src = 'comment')  AS comments_out,
        COUNT(*) FILTER (WHERE src = 'post')     AS posts_out
    FROM (
        SELECT user_id, 'like' AS src FROM likes
        WHERE created_at > (SELECT ts FROM window_start)
        UNION ALL
        SELECT user_id, 'bookmark' FROM bookmarks
        WHERE created_at > (SELECT ts FROM window_start)
        UNION ALL
        SELECT user_id, 'comment' FROM threads
        WHERE type = 'REPLY' AND created_at > (SELECT ts FROM window_start)
        UNION ALL
        SELECT user_id, 'post' FROM threads
        WHERE type = 'POST' AND created_at > (SELECT ts FROM window_start)
    ) u
    GROUP BY user_id
)
SELECT u.user_id,
    COALESCE(rc.thread_count, 0)  AS thread_count,
    COALESCE(rc.likes_in, 0)      AS likes_in,
    COALESCE(rc.comments_in, 0)   AS comments_in,
    COALESCE(rc.bookmarks_in, 0)  AS bookmarks_in,
    COALESCE(rc.reposts_in, 0)    AS reposts_in,
    COALESCE(gv.likes_out, 0)     AS likes_out,
    COALESCE(gv.comments_out, 0)  AS comments_out,
    COALESCE(gv.bookmarks_out, 0) AS bookmarks_out,
    COALESCE(gv.posts_out, 0)     AS posts_out,
    (
        COALESCE(rc.likes_in, 0)
        + 3 * COALESCE(rc.comments_in, 0)
        + 4 * COALESCE(rc.bookmarks_in, 0)
        + 5 * COALESCE(rc.reposts_in, 0)
        + 0.2 * (
            COALESCE(gv.likes_out, 0)
            + 3 * COALESCE(gv.comments_out, 0)
            + 4 * COALESCE(gv.bookmarks_out, 0)
            + COALESCE(gv.posts_out, 0)
        )
    ) AS score
FROM users u
LEFT JOIN received   rc ON rc.user_id = u.user_id
LEFT JOIN given_acts gv ON gv.user_id = u.user_id
WHERE (rc.user_id IS NOT NULL OR gv.user_id IS NOT NULL)
ORDER BY score DESC, u.user_id ASC
`

const trendingUserCountSQL = `
WITH window_start AS (SELECT NOW() - INTERVAL '7 days' AS ts),
active_users AS (
    SELECT user_id FROM threads
    WHERE created_at > (SELECT ts FROM window_start)
    UNION
    SELECT user_id FROM likes
    WHERE created_at > (SELECT ts FROM window_start)
    UNION
    SELECT user_id FROM bookmarks
    WHERE created_at > (SELECT ts FROM window_start)
    UNION
    SELECT user_id FROM reposts
    WHERE created_at > (SELECT ts FROM window_start)
    UNION
    SELECT t.user_id FROM threads t
    JOIN likes l ON l.thread_id = t.thread_id
    WHERE l.created_at > (SELECT ts FROM window_start)
    UNION
    SELECT t.user_id FROM threads t
    JOIN bookmarks b ON b.thread_id = t.thread_id
    WHERE b.created_at > (SELECT ts FROM window_start)
    UNION
    SELECT t.user_id FROM threads t
    JOIN reposts rp ON rp.thread_id = t.thread_id
    WHERE rp.created_at > (SELECT ts FROM window_start)
)
SELECT COUNT(*)
FROM active_users a
JOIN users u ON u.user_id = a.user_id
`

const userInfoSQL = `
SELECT u.user_id,
    u.user_name AS username,
    u.name,
    u.is_premium,
    COALESCE(up.avatar_object_name, '') AS avatar_object_name
FROM users u
LEFT JOIN user_profiles up ON up.user_id = u.user_id
WHERE u.user_id IN ?
`

// fetchBlockedSet returns the set of user ids in a mutual block relationship
// with the viewer, so trending lists hide blocked users (mutual hide). Empty for
// guests (viewerID 0).
func (r *trendingUserRepository) fetchBlockedSet(ctx context.Context, viewerID uint) map[uint]bool {
	set := map[uint]bool{}
	if viewerID == 0 {
		return set
	}
	var ids []uint
	r.readDB.WithContext(ctx).
		Raw(`SELECT blocked_id FROM user_blocks WHERE blocker_id = ?
			UNION
			SELECT blocker_id FROM user_blocks WHERE blocked_id = ?`, viewerID, viewerID).
		Scan(&ids)
	for _, id := range ids {
		set[id] = true
	}
	return set
}

func (r *trendingUserRepository) GetWeeklyTrending(ctx context.Context, viewerID uint, page, limit int) (trendinguserrepoiface.TrendingUsersPage, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	scores, total, cachedAt, err := r.loadScores(ctx, page, limit, offset)
	if err != nil {
		return trendinguserrepoiface.TrendingUsersPage{}, err
	}
	if len(scores) == 0 {
		return trendinguserrepoiface.TrendingUsersPage{
			Users:     []trendingusermodel.TrendingUser{},
			Total:     total,
			CachedAt:  cachedAt,
			RefreshAt: cachedAt.Add(trendingUserTTL),
		}, nil
	}

	userIDs := make([]uint, len(scores))
	for i, s := range scores {
		userIDs[i] = s.UserID
	}

	infos, err := r.fetchUserInfos(ctx, userIDs)
	if err != nil {
		return trendinguserrepoiface.TrendingUsersPage{}, err
	}
	followingSet, err := r.fetchFollowingSet(ctx, viewerID, userIDs)
	if err != nil {
		return trendinguserrepoiface.TrendingUsersPage{}, err
	}
	blockedSet := r.fetchBlockedSet(ctx, viewerID)

	users := make([]trendingusermodel.TrendingUser, 0, len(scores))
	for _, s := range scores {
		if blockedSet[s.UserID] {
			continue
		}
		info, ok := infos[s.UserID]
		if !ok {
			continue
		}
		users = append(users, trendingusermodel.TrendingUser{
			UserID:           s.UserID,
			Username:         info.Username,
			Name:             info.Name,
			AvatarObjectName: info.AvatarObjectName,
			IsPremium:        info.IsPremium,
			IsFollowing:      followingSet[s.UserID],
			Score:            s.Score,
			ThreadCount:      s.ThreadCount,
			LikesIn:          s.LikesIn,
			CommentsIn:       s.CommentsIn,
			BookmarksIn:      s.BookmarksIn,
			RepostsIn:        s.RepostsIn,
			LikesOut:         s.LikesOut,
			CommentsOut:      s.CommentsOut,
			BookmarksOut:     s.BookmarksOut,
			PostsOut:         s.PostsOut,
		})
	}

	return trendinguserrepoiface.TrendingUsersPage{
		Users:     users,
		Total:     total,
		CachedAt:  cachedAt,
		RefreshAt: cachedAt.Add(trendingUserTTL),
	}, nil
}

func (r *trendingUserRepository) loadScores(ctx context.Context, page, limit, offset int) ([]scoreRow, int64, time.Time, error) {
	key := scoresCacheKey(page, limit)
	if r.cache != nil {
		var cached cachedTrendingScores
		if err := r.cache.Get(ctx, key, &cached); err == nil {
			return cached.Rows, cached.Total, cached.CachedAt, nil
		}
	}

	var rows []scoreRow
	if err := r.readDB.WithContext(ctx).
		Raw(trendingUserScoresSQL, limit, offset).
		Scan(&rows).Error; err != nil {
		return nil, 0, time.Time{}, err
	}

	var total int64
	if err := r.readDB.WithContext(ctx).
		Raw(trendingUserCountSQL).
		Scan(&total).Error; err != nil {
		return nil, 0, time.Time{}, err
	}

	cachedAt := time.Now().UTC()
	if r.cache != nil {
		_ = r.cache.Set(ctx, key, cachedTrendingScores{Rows: rows, Total: total, CachedAt: cachedAt}, trendingUserTTL)
	}
	return rows, total, cachedAt, nil
}

func (r *trendingUserRepository) fetchUserInfos(ctx context.Context, userIDs []uint) (map[uint]userInfoRow, error) {
	var rows []userInfoRow
	if err := r.readDB.WithContext(ctx).
		Raw(userInfoSQL, userIDs).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[uint]userInfoRow, len(rows))
	for _, row := range rows {
		out[row.UserID] = row
	}
	return out, nil
}

func (r *trendingUserRepository) fetchFollowingSet(ctx context.Context, viewerID uint, userIDs []uint) (map[uint]bool, error) {
	if viewerID == 0 || len(userIDs) == 0 {
		return map[uint]bool{}, nil
	}
	var ids []uint
	if err := r.readDB.WithContext(ctx).
		Table("user_relationships").
		Where("follower_id = ? AND following_id IN ?", viewerID, userIDs).
		Pluck("following_id", &ids).Error; err != nil {
		return nil, err
	}
	out := make(map[uint]bool, len(ids))
	for _, id := range ids {
		out[id] = true
	}
	return out, nil
}

// GetMyWeeklyTrending returns the viewer's own row + 1-based rank (if active in the window)
// plus the page metadata. Reuses the cached full-ranked scores set so it stays consistent
// with the leaderboard list and avoids redundant aggregate runs.
func (r *trendingUserRepository) GetMyWeeklyTrending(ctx context.Context, viewerID uint) (*trendingusermodel.TrendingUser, *int, trendinguserrepoiface.TrendingUsersPage, error) {
	if viewerID == 0 {
		return nil, nil, trendinguserrepoiface.TrendingUsersPage{}, nil
	}
	// Wide limit so most viewers fall inside the cached page. Cache key is keyed
	// on (page, limit) so this shares a Redis entry across viewers.
	const wideLimit = 1000
	scores, total, cachedAt, err := r.loadScores(ctx, 1, wideLimit, 0)
	if err != nil {
		return nil, nil, trendinguserrepoiface.TrendingUsersPage{}, err
	}
	page := trendinguserrepoiface.TrendingUsersPage{
		Users:     nil,
		Total:     total,
		CachedAt:  cachedAt,
		RefreshAt: cachedAt.Add(trendingUserTTL),
	}
	var mine *scoreRow
	var rank int
	for i := range scores {
		if scores[i].UserID == viewerID {
			mine = &scores[i]
			rank = i + 1
			break
		}
	}
	if mine == nil {
		return nil, nil, page, nil
	}
	infos, err := r.fetchUserInfos(ctx, []uint{viewerID})
	if err != nil {
		return nil, nil, page, err
	}
	info := infos[viewerID]
	user := &trendingusermodel.TrendingUser{
		UserID:           mine.UserID,
		Username:         info.Username,
		Name:             info.Name,
		AvatarObjectName: info.AvatarObjectName,
		IsPremium:        info.IsPremium,
		IsFollowing:      false,
		Score:            mine.Score,
		ThreadCount:      mine.ThreadCount,
		LikesIn:          mine.LikesIn,
		CommentsIn:       mine.CommentsIn,
		BookmarksIn:      mine.BookmarksIn,
		RepostsIn:        mine.RepostsIn,
		LikesOut:         mine.LikesOut,
		CommentsOut:      mine.CommentsOut,
		BookmarksOut:     mine.BookmarksOut,
		PostsOut:         mine.PostsOut,
	}
	return user, &rank, page, nil
}

// ComputeAllWeeklyScores runs the full ranked aggregate (no LIMIT, no cache, no viewer info).
// Used by the snapshot scheduler — bypasses the cache so each run sees fresh data.
func (r *trendingUserRepository) ComputeAllWeeklyScores(ctx context.Context) ([]trendinguserrepoiface.WeeklyScoreRow, error) {
	var rows []scoreRow
	if err := r.readDB.WithContext(ctx).
		Raw(trendingUserAllScoresSQL).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]trendinguserrepoiface.WeeklyScoreRow, len(rows))
	for i, s := range rows {
		out[i] = trendinguserrepoiface.WeeklyScoreRow{
			UserID:       s.UserID,
			Score:        s.Score,
			ThreadCount:  s.ThreadCount,
			LikesIn:      s.LikesIn,
			CommentsIn:   s.CommentsIn,
			BookmarksIn:  s.BookmarksIn,
			RepostsIn:    s.RepostsIn,
			LikesOut:     s.LikesOut,
			CommentsOut:  s.CommentsOut,
			BookmarksOut: s.BookmarksOut,
			PostsOut:     s.PostsOut,
		}
	}
	return out, nil
}
