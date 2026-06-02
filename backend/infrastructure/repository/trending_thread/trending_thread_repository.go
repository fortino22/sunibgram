package trendingthreadrepo

import (
	"context"
	"time"

	"gorm.io/gorm"

	trendingthreadmodel "github.com/sunibgram/domain/model/trending_thread"
	trendingthreadrepoiface "github.com/sunibgram/domain/repository/trending_thread"
	"github.com/sunibgram/infrastructure/cache"
)

const (
	trendingThreadTTL = 5 * time.Minute
	// trendingThreadCandidatePool is the viewer-agnostic pool cached per tick.
	// We cache more than the top 3 so per-viewer block/mute filtering can still
	// yield 3 visible threads without re-running the aggregate.
	trendingThreadCandidatePool = 10
	trendingThreadScoresKey     = "trending:threads:weekly:top"
)

type trendingThreadRepository struct {
	readDB *gorm.DB
	cache  *cache.Cache
}

func NewTrendingThreadRepository(readDB *gorm.DB, c *cache.Cache) trendingthreadrepoiface.ITrendingThreadRepository {
	return &trendingThreadRepository{readDB: readDB, cache: c}
}

// scoreRow is the cached, viewer-agnostic aggregate per thread. Mutable fields
// (content, author name/username/avatar/is_premium) are hydrated fresh on each
// request so edits/renames are never served stale. JSON tags keep the cached
// blob compact; GORM scans by field name (snake_case), matching the SQL aliases.
type scoreRow struct {
	ThreadID     uint    `json:"t"`
	UserID       uint    `json:"u"`
	Score        float64 `json:"s"`
	LikesIn      int     `json:"li"`
	CommentsIn   int     `json:"ci"`
	BookmarksIn  int     `json:"bi"`
	RepostsIn    int     `json:"ri"`
	LikesOut     int     `json:"lo"`
	CommentsOut  int     `json:"co"`
	BookmarksOut int     `json:"bo"`
	PostsOut     int     `json:"po"`
}

type cachedScores struct {
	Rows     []scoreRow `json:"rows"`
	CachedAt time.Time  `json:"cached_at"`
}

type threadInfoRow struct {
	ThreadID         uint
	Content          string
	CreatedAt        time.Time
	UserID           uint
	Username         string
	Name             string
	IsPremium        bool
	AvatarObjectName string
}

// trendingThreadScoresSQL ranks POST threads by the engagement they RECEIVED in
// the rolling 7-day window (windowed by each interaction's created_at, so a
// thread getting fresh engagement still trends). given_acts is the same per-user
// activity CTE used by trending_user, LEFT-JOINed on author so each thread also
// carries its author's given activity. Visibility mirrors the global-trending
// scheduler: public/no community + non-private author.
const trendingThreadScoresSQL = `
WITH window_start AS (SELECT NOW() - INTERVAL '7 days' AS ts),
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
),
thread_received AS (
    SELECT t.thread_id, t.user_id,
        COALESCE(l.cnt, 0)  AS likes_in,
        COALESCE(b.cnt, 0)  AS bookmarks_in,
        COALESCE(rp.cnt, 0) AS reposts_in,
        COALESCE(cm.cnt, 0) AS comments_in
    FROM threads t
    LEFT JOIN communities c    ON c.community_id = t.community_id
    LEFT JOIN user_settings us ON us.user_id = t.user_id
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
    WHERE t.type = 'POST'
      AND (t.community_id IS NULL OR c.visibility = 'PUBLIC')
      AND COALESCE(us.is_private, false) = false
      AND (l.cnt IS NOT NULL OR b.cnt IS NOT NULL OR rp.cnt IS NOT NULL OR cm.cnt IS NOT NULL)
)
SELECT tr.thread_id,
    tr.user_id,
    tr.likes_in,
    tr.comments_in,
    tr.bookmarks_in,
    tr.reposts_in,
    COALESCE(gv.likes_out, 0)     AS likes_out,
    COALESCE(gv.comments_out, 0)  AS comments_out,
    COALESCE(gv.bookmarks_out, 0) AS bookmarks_out,
    COALESCE(gv.posts_out, 0)     AS posts_out,
    (
        tr.likes_in
        + 3 * tr.comments_in
        + 4 * tr.bookmarks_in
        + 5 * tr.reposts_in
    ) AS score
FROM thread_received tr
LEFT JOIN given_acts gv ON gv.user_id = tr.user_id
ORDER BY score DESC, tr.thread_id ASC
LIMIT ?
`

const threadInfoSQL = `
SELECT t.thread_id,
    t.content,
    t.created_at,
    t.user_id,
    u.user_name AS username,
    u.name,
    u.is_premium,
    COALESCE(up.avatar_object_name, '') AS avatar_object_name
FROM threads t
JOIN users u ON u.user_id = t.user_id
LEFT JOIN user_profiles up ON up.user_id = t.user_id
WHERE t.thread_id IN ?
`

func (r *trendingThreadRepository) GetTopWeekly(ctx context.Context, viewerID uint, limit int) (trendingthreadrepoiface.TrendingThreadsResult, error) {
	if limit < 1 {
		limit = 3
	}

	scores, cachedAt, err := r.loadScores(ctx)
	if err != nil {
		return trendingthreadrepoiface.TrendingThreadsResult{}, err
	}

	result := trendingthreadrepoiface.TrendingThreadsResult{
		Threads:   []trendingthreadmodel.TrendingThread{},
		CachedAt:  cachedAt,
		RefreshAt: cachedAt.Add(trendingThreadTTL),
	}
	if len(scores) == 0 {
		return result, nil
	}

	blockedSet := r.fetchBlockedSet(ctx, viewerID)

	// Walk the score-ordered pool, drop blocked/muted authors, take the top `limit`.
	picked := make([]scoreRow, 0, limit)
	for _, s := range scores {
		if blockedSet[s.UserID] {
			continue
		}
		picked = append(picked, s)
		if len(picked) >= limit {
			break
		}
	}
	if len(picked) == 0 {
		return result, nil
	}

	threadIDs := make([]uint, len(picked))
	for i, s := range picked {
		threadIDs[i] = s.ThreadID
	}
	infos, err := r.fetchThreadInfos(ctx, threadIDs)
	if err != nil {
		return trendingthreadrepoiface.TrendingThreadsResult{}, err
	}

	threads := make([]trendingthreadmodel.TrendingThread, 0, len(picked))
	for _, s := range picked {
		info, ok := infos[s.ThreadID]
		if !ok {
			continue
		}
		threads = append(threads, trendingthreadmodel.TrendingThread{
			ThreadID:         s.ThreadID,
			Content:          info.Content,
			CreatedAt:        info.CreatedAt,
			Score:            s.Score,
			AuthorID:         s.UserID,
			Username:         info.Username,
			Name:             info.Name,
			AvatarObjectName: info.AvatarObjectName,
			IsPremium:        info.IsPremium,
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

	result.Threads = threads
	return result, nil
}

func (r *trendingThreadRepository) loadScores(ctx context.Context) ([]scoreRow, time.Time, error) {
	if r.cache != nil {
		var cached cachedScores
		if err := r.cache.Get(ctx, trendingThreadScoresKey, &cached); err == nil {
			return cached.Rows, cached.CachedAt, nil
		}
	}

	var rows []scoreRow
	if err := r.readDB.WithContext(ctx).
		Raw(trendingThreadScoresSQL, trendingThreadCandidatePool).
		Scan(&rows).Error; err != nil {
		return nil, time.Time{}, err
	}

	cachedAt := time.Now().UTC()
	if r.cache != nil {
		_ = r.cache.Set(ctx, trendingThreadScoresKey, cachedScores{Rows: rows, CachedAt: cachedAt}, trendingThreadTTL)
	}
	return rows, cachedAt, nil
}

func (r *trendingThreadRepository) fetchThreadInfos(ctx context.Context, threadIDs []uint) (map[uint]threadInfoRow, error) {
	var rows []threadInfoRow
	if err := r.readDB.WithContext(ctx).
		Raw(threadInfoSQL, threadIDs).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[uint]threadInfoRow, len(rows))
	for _, row := range rows {
		out[row.ThreadID] = row
	}
	return out, nil
}

// fetchBlockedSet returns user ids in a mutual block or mute relationship with
// the viewer, so trending threads hide blocked/muted authors. Empty for guests.
func (r *trendingThreadRepository) fetchBlockedSet(ctx context.Context, viewerID uint) map[uint]bool {
	set := map[uint]bool{}
	if viewerID == 0 {
		return set
	}
	var ids []uint
	r.readDB.WithContext(ctx).
		Raw(`SELECT blocked_id FROM user_blocks WHERE blocker_id = ?
			UNION
			SELECT blocker_id FROM user_blocks WHERE blocked_id = ?
			UNION
			SELECT muted_user_id FROM notification_user_mutes WHERE user_id = ?`, viewerID, viewerID, viewerID).
		Scan(&ids)
	for _, id := range ids {
		set[id] = true
	}
	return set
}
