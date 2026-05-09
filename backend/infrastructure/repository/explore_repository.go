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

const exploreTTL = 60 * time.Second

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

func NewExploreRepository(readDB *gorm.DB, c *cache.Cache) domainrepo.IExploreRepository {
	return &exploreRepository{readDB: readDB, cache: c, ttl: exploreTTL}
}

func (r *exploreRepository) ExploreSearch(ctx context.Context, query string, requesterID *uint, limit int) (*model.ExploreResult, error) {
	if limit < 1 || limit > 100 {
		limit = defaultRepositoryExploreLimit
	}

	key := exploreKey(query, requesterID, limit)
	var result model.ExploreResult
	cacheable := requesterID == nil && r.cache != nil
	if cacheable {
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
	if cacheable {
		r.cache.Set(ctx, key, result, r.ttl)
	}
	return &result, nil
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
		Select("threads.user_id, threads.content").
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeReply)}).
		Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
		Where("threads.content ILIKE ? ESCAPE '\\'", pattern).
		Order("threads.created_at DESC").
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

func exploreKey(query string, requesterID *uint, limit int) string {
	viewer := "guest"
	if requesterID != nil {
		viewer = fmt.Sprintf("%d", *requesterID)
	}
	sum := sha1.Sum([]byte(strings.ToLower(strings.TrimSpace(query))))
	return fmt.Sprintf("explore:%s:viewer:%s:limit:%d", hex.EncodeToString(sum[:]), viewer, limit)
}

func likePattern(query string) string {
	query = strings.TrimSpace(query)
	query = strings.ReplaceAll(query, `\`, `\\`)
	query = strings.ReplaceAll(query, `%`, `\%`)
	query = strings.ReplaceAll(query, `_`, `\_`)
	return "%" + query + "%"
}
