package explorerepo

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	communitymodel "github.com/sunibgram/domain/model/community"
	exploremodel "github.com/sunibgram/domain/model/explore"
	threadmodel "github.com/sunibgram/domain/model/thread"
	explorerepoiface "github.com/sunibgram/domain/repository/explore"
	"github.com/sunibgram/infrastructure/cache"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
	threadrepo "github.com/sunibgram/infrastructure/repository/thread"
)

const (
	exploreTTL       = 60 * time.Second
	exploreSearchTTL = 30 * time.Second
)

func normalizeExplorePage(page int) int {
	if page < 1 {
		return 1
	}
	if page > 10000 {
		return 10000
	}
	return page
}

type exploreRepository struct {
	readDB *gorm.DB
	cache  *cache.Cache
	ttl    time.Duration
}

type ExploreUserRow struct {
	UserID           uint
	Name             string
	Username         string
	Bio              string
	IsPremium        bool
	AvatarObjectName string
	IsFollowing      bool
}

type exploreThreadRow struct {
	UserID  uint
	Content string
}

type cachedExploreThreads struct {
	Threads []threadmodel.Thread
	Total   int64
}

type cachedExploreCommunities struct {
	Communities []exploremodel.ExploreCommunityResult
	Total       int64
}

func NewExploreRepository(readDB *gorm.DB, c *cache.Cache) explorerepoiface.IExploreRepository {
	return &exploreRepository{readDB: readDB, cache: c, ttl: exploreTTL}
}

func (r *exploreRepository) ExploreSearch(ctx context.Context, query string, requesterID *uint, page, limit int) (*exploremodel.ExploreResult, error) {
	if limit < 1 || limit > 100 {
		limit = DefaultRepositoryExploreLimit
	}
	page = normalizeExplorePage(page)
	offset := (page - 1) * limit

	key := exploreSearchKey(query, requesterID, page, limit)
	var result exploremodel.ExploreResult
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return &result, nil
		}
	}

	users, err := r.exploreUsers(ctx, query, requesterID, offset, limit)
	if err != nil {
		return nil, err
	}
	threads, err := r.exploreThreads(ctx, query, offset, limit)
	if err != nil {
		return nil, err
	}

	result = exploremodel.ExploreResult{Users: users, Threads: threads}
	if r.cache != nil {
		if err := r.cache.Set(ctx, key, result, exploreSearchTTL); err == nil {
			r.cache.RegisterCacheKey(ctx, infrarepoShared.ExploreGlobalTag(), key)
		}
	}
	return &result, nil
}

func (r *exploreRepository) FindTopPosts(ctx context.Context, query string, requesterID uint, page, limit int) ([]threadmodel.Thread, int64, error) {
	return r.findExploreThreads(ctx, "top", query, requesterID, []string{string(threadmodel.ThreadTypePost), string(threadmodel.ThreadTypeReply), string(threadmodel.ThreadTypeRepost)}, page, limit, false, "(threads.like_count + threads.comment_count + threads.repost_count) DESC, threads.created_at DESC")
}

func (r *exploreRepository) FindLatestPosts(ctx context.Context, query string, requesterID uint, page, limit int) ([]threadmodel.Thread, int64, error) {
	return r.findExploreThreads(ctx, "latest", query, requesterID, []string{string(threadmodel.ThreadTypePost), string(threadmodel.ThreadTypeReply), string(threadmodel.ThreadTypeRepost)}, page, limit, false, "threads.created_at DESC")
}

func (r *exploreRepository) FindPeople(ctx context.Context, query string, requesterID uint, page, limit int) ([]exploremodel.ExploreUserResult, error) {
	if limit < 1 || limit > 100 {
		limit = DefaultRepositoryExploreLimit
	}
	page = normalizeExplorePage(page)
	offset := (page - 1) * limit

	key := fmt.Sprintf("explore:people:%s:viewer:%d:page:%d:limit:%d", exploreQueryHash(query), requesterID, page, limit)
	var users []exploremodel.ExploreUserResult
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &users); err == nil {
			return users, nil
		}
	}

	users, err := r.exploreUsers(ctx, query, &requesterID, offset, limit)
	if err != nil {
		return nil, err
	}
	if r.cache != nil {
		if err := r.cache.Set(ctx, key, users, r.ttl); err == nil {
			r.cache.RegisterCacheKey(ctx, infrarepoShared.ExploreGlobalTag(), key)
		}
	}
	return users, nil
}

func (r *exploreRepository) FindMediaPosts(ctx context.Context, query string, requesterID uint, page, limit int) ([]threadmodel.Thread, int64, error) {
	return r.findExploreThreads(ctx, "media", query, requesterID, []string{string(threadmodel.ThreadTypePost), string(threadmodel.ThreadTypeReply)}, page, limit, true, "MAX(threads.created_at) DESC")
}

func (r *exploreRepository) FindCommunities(ctx context.Context, query string, requesterID uint, page, limit int) ([]exploremodel.ExploreCommunityResult, int64, error) {
	if limit < 1 || limit > 100 {
		limit = DefaultRepositoryExploreLimit
	}
	page = normalizeExplorePage(page)
	offset := (page - 1) * limit

	key := fmt.Sprintf("explore:communities:%s:viewer:%d:page:%d:limit:%d", exploreQueryHash(query), requesterID, page, limit)
	var result cachedExploreCommunities
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return result.Communities, result.Total, nil
		}
	}

	pattern := LikePattern(query)
	base := r.readDB.WithContext(ctx).
		Model(&communitymodel.Community{}).
		Joins("LEFT JOIN categories ON categories.id = communities.category_id").
		Where("communities.name ILIKE ? ESCAPE '\\' OR communities.description ILIKE ? ESCAPE '\\' OR categories.name ILIKE ? ESCAPE '\\'", pattern, pattern, pattern)

	var communities []exploremodel.ExploreCommunityResult
	if err := base.
		Select(`
			communities.community_id,
			communities.name,
			communities.member_count,
			COALESCE(communities.banner_object_name, '') AS banner_object_name,
			(community_members.user_id IS NOT NULL) AS is_join
		`).
		Joins("LEFT JOIN community_members ON community_members.community_id = communities.community_id AND community_members.user_id = ?", requesterID).
		Order("communities.member_count DESC, communities.created_at DESC").
		Offset(offset).
		Limit(limit + 1).
		Scan(&communities).Error; err != nil {
		return nil, 0, err
	}

	hasMore := len(communities) > limit
	if hasMore {
		communities = communities[:limit]
	}
	total := int64(offset + len(communities))
	if hasMore {
		total++
	}

	result = cachedExploreCommunities{Communities: communities, Total: total}
	if r.cache != nil {
		if err := r.cache.Set(ctx, key, result, r.ttl); err == nil {
			r.cache.RegisterCacheKey(ctx, infrarepoShared.ExploreGlobalTag(), key)
		}
	}
	return communities, total, nil
}

// FindFacultySuggestions suggests people who share at least one faculty with the
// requester and whom the requester does not already follow. Empty if the
// requester has no faculty (or no faculty peers left to suggest).
func (r *exploreRepository) FindFacultySuggestions(ctx context.Context, requesterID uint, page, limit int) ([]exploremodel.ExploreUserResult, error) {
	limit, offset := clampExploreLimitOffset(limit, page)

	key := exploreSectionKey("suggested_faculty", "", fmt.Sprintf("%d:page:%d", requesterID, page), limit)
	if users, ok := r.getCachedSuggestions(ctx, key); ok {
		return users, nil
	}

	var rows []ExploreUserRow
	if err := r.readDB.WithContext(ctx).
		Table("user_faculties AS me").
		Select(suggestionSelectColumns).
		Joins("JOIN user_faculties AS peer ON peer.faculty_id = me.faculty_id AND peer.user_id <> me.user_id").
		Joins("JOIN users ON users.user_id = peer.user_id").
		Joins("JOIN roles ON roles.role_id = users.role_id").
		Joins("LEFT JOIN user_profiles ON user_profiles.user_id = users.user_id").
		Joins("LEFT JOIN user_relationships AS existing ON existing.follower_id = ? AND existing.following_id = peer.user_id", requesterID).
		Where("me.user_id = ?", requesterID).
		Where("existing.following_id IS NULL").
		Where("roles.name <> ?", "admin").
		Group(suggestionGroupColumns).
		Order("users.is_premium DESC, users.created_at DESC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	users := mapExploreUserRows(rows)
	r.cacheSuggestions(ctx, key, users)
	return users, nil
}

// FindMutualSuggestions ranks mutual-following candidates (friends-of-friends):
// people followed by the accounts the requester already follows, ordered by
// shared-connection count. Empty if the requester follows nobody.
func (r *exploreRepository) FindMutualSuggestions(ctx context.Context, requesterID uint, page, limit int) ([]exploremodel.ExploreUserResult, error) {
	limit, offset := clampExploreLimitOffset(limit, page)

	key := exploreSectionKey("suggested_mutual", "", fmt.Sprintf("%d:page:%d", requesterID, page), limit)
	if users, ok := r.getCachedSuggestions(ctx, key); ok {
		return users, nil
	}

	var rows []ExploreUserRow
	if err := r.readDB.WithContext(ctx).
		Table("user_relationships AS followed").
		Select(suggestionSelectColumns).
		Joins("JOIN user_relationships AS suggested ON suggested.follower_id = followed.following_id").
		Joins("JOIN users ON users.user_id = suggested.following_id").
		Joins("JOIN roles ON roles.role_id = users.role_id").
		Joins("LEFT JOIN user_profiles ON user_profiles.user_id = users.user_id").
		Joins("LEFT JOIN user_relationships AS existing ON existing.follower_id = ? AND existing.following_id = suggested.following_id", requesterID).
		Where("followed.follower_id = ?", requesterID).
		Where("suggested.following_id <> ?", requesterID).
		Where("existing.following_id IS NULL").
		Where("roles.name <> ?", "admin").
		Group(suggestionGroupColumns).
		Order("COUNT(DISTINCT followed.following_id) DESC, users.is_premium DESC, users.created_at DESC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	users := mapExploreUserRows(rows)
	r.cacheSuggestions(ctx, key, users)
	return users, nil
}

// FindRandomSuggestions returns random people on the app whom the requester does
// not already follow. Used as the last-resort fallback for the preview widget.
func (r *exploreRepository) FindRandomSuggestions(ctx context.Context, requesterID uint, page, limit int) ([]exploremodel.ExploreUserResult, error) {
	limit, offset := clampExploreLimitOffset(limit, page)

	key := exploreSectionKey("suggested_random", "", fmt.Sprintf("%d:page:%d", requesterID, page), limit)
	if users, ok := r.getCachedSuggestions(ctx, key); ok {
		return users, nil
	}

	var rows []ExploreUserRow
	if err := r.readDB.WithContext(ctx).
		Table("users").
		Select(suggestionSelectColumns).
		Joins("JOIN roles ON roles.role_id = users.role_id").
		Joins("LEFT JOIN user_profiles ON user_profiles.user_id = users.user_id").
		Joins("LEFT JOIN user_relationships AS existing ON existing.follower_id = ? AND existing.following_id = users.user_id", requesterID).
		Where("users.user_id <> ?", requesterID).
		Where("existing.following_id IS NULL").
		Where("roles.name <> ?", "admin").
		Order("RANDOM()").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	users := mapExploreUserRows(rows)
	r.cacheSuggestions(ctx, key, users)
	return users, nil
}

// suggestionSelectColumns / suggestionGroupColumns are shared by both suggestion
// queries so the projected fields (incl. bio) and GROUP BY stay in lockstep.
const suggestionSelectColumns = `
	users.user_id,
	users.name,
	users.user_name AS username,
	COALESCE(user_profiles.bio, '') AS bio,
	users.is_premium,
	COALESCE(user_profiles.avatar_object_name, '') AS avatar_object_name,
	FALSE AS is_following
`

const suggestionGroupColumns = "users.user_id, users.name, users.user_name, users.is_premium, users.created_at, user_profiles.bio, user_profiles.avatar_object_name"

func mapExploreUserRows(rows []ExploreUserRow) []exploremodel.ExploreUserResult {
	users := make([]exploremodel.ExploreUserResult, len(rows))
	for i, row := range rows {
		users[i] = exploremodel.ExploreUserResult{
			UserID:           row.UserID,
			Name:             row.Name,
			Username:         row.Username,
			Bio:              row.Bio,
			IsPremium:        row.IsPremium,
			AvatarObjectName: row.AvatarObjectName,
			IsFollowing:      row.IsFollowing,
		}
	}
	return users
}

func (r *exploreRepository) getCachedSuggestions(ctx context.Context, key string) ([]exploremodel.ExploreUserResult, bool) {
	if r.cache == nil {
		return nil, false
	}
	var users []exploremodel.ExploreUserResult
	if err := r.cache.Get(ctx, key, &users); err == nil {
		return users, true
	}
	return nil, false
}

func (r *exploreRepository) cacheSuggestions(ctx context.Context, key string, users []exploremodel.ExploreUserResult) {
	if r.cache == nil {
		return
	}
	if err := r.cache.Set(ctx, key, users, r.ttl); err == nil {
		r.cache.RegisterCacheKey(ctx, infrarepoShared.ExploreGlobalTag(), key)
	}
}

// clampExploreLimitOffset normalizes the limit/page pair and returns the
// effective limit and DB offset.
func clampExploreLimitOffset(limit, page int) (int, int) {
	if limit < 1 || limit > 100 {
		limit = DefaultRepositoryExploreLimit
	}
	page = normalizeExplorePage(page)
	return limit, (page - 1) * limit
}

func (r *exploreRepository) findExploreThreads(ctx context.Context, section, query string, requesterID uint, types []string, page, limit int, mediaOnly bool, order string) ([]threadmodel.Thread, int64, error) {
	if limit < 1 || limit > 100 {
		limit = DefaultRepositoryExploreLimit
	}
	page = normalizeExplorePage(page)
	offset := (page - 1) * limit

	key := exploreThreadSectionKey(section, query, requesterID, page, limit)
	var result cachedExploreThreads
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return result.Threads, result.Total, nil
		}
	}

	base := r.exploreThreadBase(ctx, query, &requesterID, types)
	if mediaOnly {
		base = base.Joins("JOIN thread_media ON thread_media.thread_id = threads.thread_id")
	}

	var total int64
	var countQuery *gorm.DB
	if mediaOnly {
		countQuery = base.Select("COUNT(DISTINCT threads.thread_id)")
	} else {
		countQuery = base
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var ids []uint
	var idQuery *gorm.DB
	if mediaOnly {
		idQuery = base.Select("threads.thread_id").Group("threads.thread_id").Order("MAX(threads.created_at) DESC")
	} else {
		idQuery = base.Select("threads.thread_id").Order(order)
	}
	if err := idQuery.Offset(offset).Limit(limit).Pluck("thread_id", &ids).Error; err != nil {
		return nil, 0, err
	}
	if len(ids) == 0 {
		result = cachedExploreThreads{Threads: []threadmodel.Thread{}, Total: total}
		if r.cache != nil {
			if err := r.cache.Set(ctx, key, result, r.ttl); err == nil {
				r.registerExploreSectionTags(ctx, section, key)
			}
		}
		return result.Threads, result.Total, nil
	}

	var threads []threadmodel.Thread
	var err error
	if mediaOnly {
		threads, err = r.loadThreadSummariesByIDs(ctx, ids)
	} else {
		threads, err = r.loadThreadsByIDs(ctx, ids)
	}
	if err != nil {
		return nil, 0, err
	}
	result = cachedExploreThreads{Threads: threads, Total: total}
	if r.cache != nil {
		if err := r.cache.Set(ctx, key, result, r.ttl); err == nil {
			r.registerExploreSectionTags(ctx, section, key)
		}
	}
	return result.Threads, result.Total, nil
}

// registerExploreSectionTags wires per-section tags + global. Top/latest get
// their dedicated tag so InvalidateExploreTopCache can drop just top without
// touching latest, search, people, communities.
func (r *exploreRepository) registerExploreSectionTags(ctx context.Context, section, key string) {
	r.cache.RegisterCacheKey(ctx, infrarepoShared.ExploreGlobalTag(), key)
	switch section {
	case "top":
		r.cache.RegisterCacheKey(ctx, infrarepoShared.ExploreTopTag(), key)
	case "latest":
		r.cache.RegisterCacheKey(ctx, infrarepoShared.ExploreLatestTag(), key)
	}
}

func (r *exploreRepository) exploreThreadBase(ctx context.Context, query string, requesterID *uint, types []string) *gorm.DB {
	pattern := LikePattern(query)
	db := r.readDB.WithContext(ctx).
		Model(&threadmodel.Thread{}).
		Joins("LEFT JOIN threads original ON original.thread_id = threads.original_thread_id").
		Joins("LEFT JOIN communities ON communities.community_id = COALESCE(threads.community_id, original.community_id)").
		Where("threads.type IN ?", types).
		Where("COALESCE(threads.community_id, original.community_id) IS NULL OR communities.visibility = ?", communitymodel.CommunityVisibilityPublic).
		Where("COALESCE(NULLIF(threads.content, ''), original.content) ILIKE ? ESCAPE '\\'", pattern)
	return threadrepo.ApplyThreadVisibility(db, requesterID)
}

func (r *exploreRepository) loadThreadsByIDs(ctx context.Context, ids []uint) ([]threadmodel.Thread, error) {
	var rows []threadmodel.Thread
	if err := threadrepo.PreloadThreadDetails(r.readDB.WithContext(ctx)).
		Where("threads.thread_id IN ?", ids).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	byID := make(map[uint]threadmodel.Thread, len(rows))
	for _, row := range rows {
		byID[row.ThreadID] = row
	}

	threads := make([]threadmodel.Thread, 0, len(ids))
	for _, id := range ids {
		if thread, ok := byID[id]; ok {
			threads = append(threads, thread)
		}
	}
	return threads, nil
}

func (r *exploreRepository) loadThreadSummariesByIDs(ctx context.Context, ids []uint) ([]threadmodel.Thread, error) {
	var rows []threadmodel.Thread
	if err := r.readDB.WithContext(ctx).
		Select("thread_id", "user_id").
		Where("threads.thread_id IN ?", ids).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	byID := make(map[uint]threadmodel.Thread, len(rows))
	for _, row := range rows {
		byID[row.ThreadID] = row
	}

	threads := make([]threadmodel.Thread, 0, len(ids))
	for _, id := range ids {
		if thread, ok := byID[id]; ok {
			threads = append(threads, thread)
		}
	}
	return threads, nil
}

const DefaultRepositoryExploreLimit = 20

func (r *exploreRepository) exploreUsers(ctx context.Context, query string, requesterID *uint, offset, limit int) ([]exploremodel.ExploreUserResult, error) {
	pattern := LikePattern(query)
	viewerID := uint(0)
	if requesterID != nil {
		viewerID = *requesterID
	}

	var rows []ExploreUserRow
	if err := r.readDB.WithContext(ctx).
		Table("users").
		Select(`
			users.user_id,
			users.name,
			users.user_name AS username,
			users.is_premium,
			COALESCE(user_profiles.avatar_object_name, '') AS avatar_object_name,
			(user_relationships.following_id IS NOT NULL) AS is_following
		`).
		Joins("JOIN roles ON roles.role_id = users.role_id").
		Joins("LEFT JOIN user_profiles ON user_profiles.user_id = users.user_id").
		Joins("LEFT JOIN user_relationships ON user_relationships.following_id = users.user_id AND user_relationships.follower_id = ?", viewerID).
		Where("roles.name <> ?", "admin").
		Where("users.user_name ILIKE ? ESCAPE '\\' OR users.name ILIKE ? ESCAPE '\\' OR user_profiles.bio ILIKE ? ESCAPE '\\'", pattern, pattern, pattern).
		Order("users.is_premium DESC, users.created_at DESC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	users := make([]exploremodel.ExploreUserResult, len(rows))
	for i, row := range rows {
		users[i] = exploremodel.ExploreUserResult{
			UserID:           row.UserID,
			Name:             row.Name,
			Username:         row.Username,
			IsPremium:        row.IsPremium,
			AvatarObjectName: row.AvatarObjectName,
			IsFollowing:      row.IsFollowing,
		}
	}
	return users, nil
}

func (r *exploreRepository) exploreThreads(ctx context.Context, query string, offset, limit int) ([]exploremodel.ExploreThreadResult, error) {
	pattern := LikePattern(query)
	var rows []exploreThreadRow
	if err := r.readDB.WithContext(ctx).
		Table("threads").
		Select("DISTINCT ON (threads.content) threads.user_id, threads.content").
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN user_settings thread_visibility_settings ON thread_visibility_settings.user_id = threads.user_id").
		Where("threads.type IN ?", []string{string(threadmodel.ThreadTypePost), string(threadmodel.ThreadTypeReply)}).
		Where("threads.community_id IS NULL OR communities.visibility = ?", communitymodel.CommunityVisibilityPublic).
		Where("COALESCE(thread_visibility_settings.is_private, false) = ?", false).
		Where("threads.content ILIKE ? ESCAPE '\\'", pattern).
		Order("threads.content, threads.created_at DESC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	threads := make([]exploremodel.ExploreThreadResult, len(rows))
	for i, row := range rows {
		threads[i] = exploremodel.ExploreThreadResult{
			UserID:  row.UserID,
			Content: row.Content,
		}
	}
	return threads, nil
}

func exploreSearchKey(query string, requesterID *uint, page, limit int) string {
	viewer := "guest"
	if requesterID != nil {
		viewer = fmt.Sprintf("%d", *requesterID)
	}
	return fmt.Sprintf("explore:search:%s:%d:%d:viewer:%s", exploreQueryHash(query), page, limit, viewer)
}

func exploreThreadSectionKey(section, query string, requesterID uint, page, limit int) string {
	return fmt.Sprintf("explore:%s:%s:%d:%d:viewer:%d", section, exploreQueryHash(query), page, limit, requesterID)
}

func exploreSectionKey(section, query, viewer string, limit int) string {
	sum := exploreQueryHash(query)
	if section == "" {
		return fmt.Sprintf("explore:%s:viewer:%s:limit:%d", sum, viewer, limit)
	}
	return fmt.Sprintf("explore:%s:%s:viewer:%s:limit:%d", section, sum, viewer, limit)
}

func exploreQueryHash(query string) string {
	sum := sha1.Sum([]byte(strings.ToLower(strings.TrimSpace(query))))
	return hex.EncodeToString(sum[:])
}

func LikePattern(query string) string {
	query = strings.TrimSpace(query)
	query = strings.ReplaceAll(query, `\`, `\\`)
	query = strings.ReplaceAll(query, `%`, `\%`)
	query = strings.ReplaceAll(query, `_`, `\_`)
	return "%" + query + "%"
}
