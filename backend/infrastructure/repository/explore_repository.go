package repository

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
)

const (
	exploreTTL       = 60 * time.Second
	exploreSearchTTL = 30 * time.Second
	explorePage      = 1
)

type exploreRepository struct {
	readDB *gorm.DB
	cache  *cache.Cache
	ttl    time.Duration
}

type exploreUserRow struct {
	UserID           uint
	Name             string
	Username         string
	IsPremium        bool
	AvatarObjectName string
	AvatarBucket     string
	IsFollowing      bool
}

type exploreThreadRow struct {
	UserID  uint
	Content string
}

type cachedExploreThreads struct {
	Threads []model.Thread
	Total   int64
}

type cachedExploreCommunities struct {
	Communities []model.ExploreCommunityResult
	Total       int64
}

func NewExploreRepository(readDB *gorm.DB, c *cache.Cache) domainrepo.IExploreRepository {
	return &exploreRepository{readDB: readDB, cache: c, ttl: exploreTTL}
}

func (r *exploreRepository) ExploreSearch(ctx context.Context, query string, requesterID *uint, limit int) (*model.ExploreResult, error) {
	if limit < 1 || limit > 100 {
		limit = defaultRepositoryExploreLimit
	}

	key := exploreSearchKey(query, requesterID, explorePage, limit)
	var result model.ExploreResult
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return &result, nil
		}
	}

	users, err := r.exploreUsers(ctx, query, requesterID, limit)
	if err != nil {
		return nil, err
	}
	threads, err := r.exploreThreads(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	result = model.ExploreResult{Users: users, Threads: threads}
	if r.cache != nil {
		_ = r.cache.Set(ctx, key, result, exploreSearchTTL)
	}
	return &result, nil
}

func (r *exploreRepository) FindTopPosts(ctx context.Context, query string, requesterID uint, limit int) ([]model.Thread, int64, error) {
	return r.findExploreThreads(ctx, "top", query, requesterID, []string{string(model.ThreadTypePost), string(model.ThreadTypeReply), string(model.ThreadTypeRepost)}, limit, false, "(threads.like_count + threads.comment_count + threads.repost_count) DESC, threads.created_at DESC")
}

func (r *exploreRepository) FindLatestPosts(ctx context.Context, query string, requesterID uint, limit int) ([]model.Thread, int64, error) {
	return r.findExploreThreads(ctx, "latest", query, requesterID, []string{string(model.ThreadTypePost), string(model.ThreadTypeReply), string(model.ThreadTypeRepost)}, limit, false, "threads.created_at DESC")
}

func (r *exploreRepository) FindPeople(ctx context.Context, query string, requesterID uint, limit int) ([]model.ExploreUserResult, error) {
	if limit < 1 || limit > 100 {
		limit = defaultRepositoryExploreLimit
	}

	key := exploreSectionKey("people", query, fmt.Sprintf("%d", requesterID), limit)
	var users []model.ExploreUserResult
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &users); err == nil {
			return users, nil
		}
	}

	users, err := r.exploreUsers(ctx, query, &requesterID, limit)
	if err != nil {
		return nil, err
	}
	if r.cache != nil {
		_ = r.cache.Set(ctx, key, users, r.ttl)
	}
	return users, nil
}

func (r *exploreRepository) FindMediaPosts(ctx context.Context, query string, requesterID uint, limit int) ([]model.Thread, int64, error) {
	return r.findExploreThreads(ctx, "media", query, requesterID, []string{string(model.ThreadTypePost), string(model.ThreadTypeReply)}, limit, true, "MAX(threads.created_at) DESC")
}

func (r *exploreRepository) FindCommunities(ctx context.Context, query string, requesterID uint, limit int) ([]model.ExploreCommunityResult, int64, error) {
	if limit < 1 || limit > 100 {
		limit = defaultRepositoryExploreLimit
	}

	key := exploreSectionKey("communities", query, fmt.Sprintf("%d", requesterID), limit)
	var result cachedExploreCommunities
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return result.Communities, result.Total, nil
		}
	}

	pattern := likePattern(query)
	base := r.readDB.WithContext(ctx).
		Model(&model.Community{}).
		Joins("LEFT JOIN categories ON categories.id = communities.category_id").
		Where("communities.name ILIKE ? ESCAPE '\\' OR communities.description ILIKE ? ESCAPE '\\' OR categories.name ILIKE ? ESCAPE '\\'", pattern, pattern, pattern)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var communities []model.ExploreCommunityResult
	if err := base.
		Select(`
			communities.community_id,
			communities.name,
			communities.member_count,
			(community_members.user_id IS NOT NULL) AS is_join
		`).
		Joins("LEFT JOIN community_members ON community_members.community_id = communities.community_id AND community_members.user_id = ?", requesterID).
		Order("communities.member_count DESC, communities.created_at DESC").
		Limit(limit).
		Scan(&communities).Error; err != nil {
		return nil, 0, err
	}

	result = cachedExploreCommunities{Communities: communities, Total: total}
	if r.cache != nil {
		_ = r.cache.Set(ctx, key, result, r.ttl)
	}
	return communities, total, nil
}

func (r *exploreRepository) FindSuggestedPeople(ctx context.Context, requesterID uint, limit int) ([]model.ExploreUserResult, error) {
	if limit < 1 || limit > 100 {
		limit = defaultRepositoryExploreLimit
	}

	key := exploreSectionKey("suggested_people", "", fmt.Sprintf("%d", requesterID), limit)
	var users []model.ExploreUserResult
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &users); err == nil {
			return users, nil
		}
	}

	var rows []exploreUserRow
	if err := r.readDB.WithContext(ctx).
		Table("user_relationships AS followed").
		Select(`
			users.user_id,
			users.name,
			users.user_name AS username,
			users.is_premium,
			COALESCE(user_profiles.avatar_object_name, '') AS avatar_object_name,
			COALESCE(user_profiles.avatar_bucket, '') AS avatar_bucket,
			FALSE AS is_following
		`).
		Joins("JOIN user_relationships AS suggested ON suggested.follower_id = followed.following_id").
		Joins("JOIN users ON users.user_id = suggested.following_id").
		Joins("JOIN roles ON roles.role_id = users.role_id").
		Joins("LEFT JOIN user_profiles ON user_profiles.user_id = users.user_id").
		Joins("LEFT JOIN user_relationships AS existing ON existing.follower_id = ? AND existing.following_id = suggested.following_id", requesterID).
		Where("followed.follower_id = ?", requesterID).
		Where("suggested.following_id <> ?", requesterID).
		Where("existing.following_id IS NULL").
		Where("roles.name <> ?", "admin").
		Group("users.user_id, users.name, users.user_name, users.is_premium, users.created_at, user_profiles.avatar_object_name, user_profiles.avatar_bucket").
		Order("COUNT(DISTINCT followed.following_id) DESC, users.is_premium DESC, users.created_at DESC").
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	users = make([]model.ExploreUserResult, len(rows))
	for i, row := range rows {
		users[i] = model.ExploreUserResult{
			UserID:           row.UserID,
			Name:             row.Name,
			Username:         row.Username,
			IsPremium:        row.IsPremium,
			AvatarObjectName: row.AvatarObjectName,
			AvatarBucket:     row.AvatarBucket,
			IsFollowing:      row.IsFollowing,
		}
	}
	if r.cache != nil {
		_ = r.cache.Set(ctx, key, users, r.ttl)
	}
	return users, nil
}

func (r *exploreRepository) findExploreThreads(ctx context.Context, section, query string, requesterID uint, types []string, limit int, mediaOnly bool, order string) ([]model.Thread, int64, error) {
	if limit < 1 || limit > 100 {
		limit = defaultRepositoryExploreLimit
	}

	key := exploreSectionKey(section, query, fmt.Sprintf("%d", requesterID), limit)
	if section == "top" || section == "latest" {
		key = exploreThreadSectionKey(section, query, requesterID, explorePage, limit)
	}
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
	if err := idQuery.Limit(limit).Pluck("thread_id", &ids).Error; err != nil {
		return nil, 0, err
	}
	if len(ids) == 0 {
		result = cachedExploreThreads{Threads: []model.Thread{}, Total: total}
		if r.cache != nil {
			_ = r.cache.Set(ctx, key, result, r.ttl)
		}
		return result.Threads, result.Total, nil
	}

	var threads []model.Thread
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
		_ = r.cache.Set(ctx, key, result, r.ttl)
	}
	return result.Threads, result.Total, nil
}

func (r *exploreRepository) exploreThreadBase(ctx context.Context, query string, requesterID *uint, types []string) *gorm.DB {
	pattern := likePattern(query)
	db := r.readDB.WithContext(ctx).
		Model(&model.Thread{}).
		Joins("LEFT JOIN threads original ON original.thread_id = threads.original_thread_id").
		Joins("LEFT JOIN communities ON communities.community_id = COALESCE(threads.community_id, original.community_id)").
		Where("threads.type IN ?", types).
		Where("COALESCE(threads.community_id, original.community_id) IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
		Where("COALESCE(NULLIF(threads.content, ''), original.content) ILIKE ? ESCAPE '\\'", pattern)
	return applyThreadVisibility(db, requesterID)
}

func (r *exploreRepository) loadThreadsByIDs(ctx context.Context, ids []uint) ([]model.Thread, error) {
	var rows []model.Thread
	if err := preloadThreadDetails(r.readDB.WithContext(ctx)).
		Where("threads.thread_id IN ?", ids).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	byID := make(map[uint]model.Thread, len(rows))
	for _, row := range rows {
		byID[row.ThreadID] = row
	}

	threads := make([]model.Thread, 0, len(ids))
	for _, id := range ids {
		if thread, ok := byID[id]; ok {
			threads = append(threads, thread)
		}
	}
	return threads, nil
}

func (r *exploreRepository) loadThreadSummariesByIDs(ctx context.Context, ids []uint) ([]model.Thread, error) {
	var rows []model.Thread
	if err := r.readDB.WithContext(ctx).
		Select("thread_id", "user_id").
		Where("threads.thread_id IN ?", ids).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	byID := make(map[uint]model.Thread, len(rows))
	for _, row := range rows {
		byID[row.ThreadID] = row
	}

	threads := make([]model.Thread, 0, len(ids))
	for _, id := range ids {
		if thread, ok := byID[id]; ok {
			threads = append(threads, thread)
		}
	}
	return threads, nil
}

const defaultRepositoryExploreLimit = 20

func (r *exploreRepository) exploreUsers(ctx context.Context, query string, requesterID *uint, limit int) ([]model.ExploreUserResult, error) {
	pattern := likePattern(query)
	viewerID := uint(0)
	if requesterID != nil {
		viewerID = *requesterID
	}

	var rows []exploreUserRow
	if err := r.readDB.WithContext(ctx).
		Table("users").
		Select(`
			users.user_id,
			users.name,
			users.user_name AS username,
			users.is_premium,
			COALESCE(user_profiles.avatar_object_name, '') AS avatar_object_name,
			COALESCE(user_profiles.avatar_bucket, '') AS avatar_bucket,
			(user_relationships.following_id IS NOT NULL) AS is_following
		`).
		Joins("JOIN roles ON roles.role_id = users.role_id").
		Joins("LEFT JOIN user_profiles ON user_profiles.user_id = users.user_id").
		Joins("LEFT JOIN user_relationships ON user_relationships.following_id = users.user_id AND user_relationships.follower_id = ?", viewerID).
		Where("roles.name <> ?", "admin").
		Where("users.user_name ILIKE ? ESCAPE '\\' OR users.name ILIKE ? ESCAPE '\\' OR user_profiles.bio ILIKE ? ESCAPE '\\'", pattern, pattern, pattern).
		Order("users.is_premium DESC, users.created_at DESC").
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	users := make([]model.ExploreUserResult, len(rows))
	for i, row := range rows {
		users[i] = model.ExploreUserResult{
			UserID:           row.UserID,
			Name:             row.Name,
			Username:         row.Username,
			IsPremium:        row.IsPremium,
			AvatarObjectName: row.AvatarObjectName,
			AvatarBucket:     row.AvatarBucket,
			IsFollowing:      row.IsFollowing,
		}
	}
	return users, nil
}

func (r *exploreRepository) exploreThreads(ctx context.Context, query string, limit int) ([]model.ExploreThreadResult, error) {
	pattern := likePattern(query)
	var rows []exploreThreadRow
	if err := r.readDB.WithContext(ctx).
		Table("threads").
		Select("DISTINCT ON (threads.content) threads.user_id, threads.content").
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN user_settings thread_visibility_settings ON thread_visibility_settings.user_id = threads.user_id").
		Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeReply)}).
		Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
		Where("COALESCE(thread_visibility_settings.is_private, false) = ?", false).
		Where("threads.content ILIKE ? ESCAPE '\\'", pattern).
		Order("threads.content, threads.created_at DESC").
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	threads := make([]model.ExploreThreadResult, len(rows))
	for i, row := range rows {
		threads[i] = model.ExploreThreadResult{
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

func likePattern(query string) string {
	query = strings.TrimSpace(query)
	query = strings.ReplaceAll(query, `\`, `\\`)
	query = strings.ReplaceAll(query, `%`, `\%`)
	query = strings.ReplaceAll(query, `_`, `\_`)
	return "%" + query + "%"
}
