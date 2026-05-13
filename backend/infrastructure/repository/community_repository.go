package repository

import (
	"context"
	"fmt"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
)

type communityRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewCommunityRepository(writeDB, readDB *gorm.DB, c *cache.Cache) domainrepo.ICommunityRepository {
	return &communityRepository{writeDB: writeDB, readDB: readDB, cache: c}
}

func communityIDKey(id uint) string {
	return fmt.Sprintf("community:id:%d", id)
}

func communityListKey(category string, page, limit int) string {
	return fmt.Sprintf("community:list:%s:page:%d:limit:%d", category, page, limit)
}

func communityCreatedKey(userID uint, page, limit int) string {
	return fmt.Sprintf("community:created:%d:page:%d:limit:%d", userID, page, limit)
}

func communityJoinedKey(userID uint, page, limit int) string {
	return fmt.Sprintf("community:joined:%d:page:%d:limit:%d", userID, page, limit)
}

type cachedCommunities struct {
	Communities []model.Community
	Total       int64
}

func (r *communityRepository) Save(community *model.Community) error {
	return r.SaveWithTx(r.writeDB, community)
}

func (r *communityRepository) SaveWithTx(tx *gorm.DB, community *model.Community) error {
	if err := tx.Create(community).Error; err != nil {
		return err
	}
	r.invalidateCommunityCreate(community.UserID)
	return nil
}

func (r *communityRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.writeDB.Transaction(fn)
}

func (r *communityRepository) FindByID(id uint) (*model.Community, error) {
	ctx := context.Background()
	key := communityIDKey(id)
	var community model.Community
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &community); err == nil {
			return &community, nil
		}
	}
	if err := r.readDB.Preload("User").First(&community, id).Error; err != nil {
		return nil, err
	}
	if r.cache != nil {
		_ = r.cache.Set(ctx, key, community, cache.DefaultTTL)
	}
	return &community, nil
}

func (r *communityRepository) Update(community *model.Community) error {
	if err := r.writeDB.Model(community).UpdateColumns(map[string]any{
		"name":               community.Name,
		"description":        community.Description,
		"rules":              community.Rules,
		"category":           community.Category,
		"visibility":         community.Visibility,
		"avatar_object_name": community.AvatarObjectName,
		"avatar_bucket":      community.AvatarBucket,
		"banner_object_name": community.BannerObjectName,
		"banner_bucket":      community.BannerBucket,
		"updated_at":         gorm.Expr("NOW()"),
	}).Error; err != nil {
		return err
	}
	r.invalidateCommunityUpdate(community.CommunityID, community.UserID)
	return nil
}

func (r *communityRepository) Delete(id uint) error {
	var community model.Community
	_ = r.writeDB.Select("community_id", "user_id").First(&community, id).Error
	if err := r.writeDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("community_id = ?", id).Delete(&model.CommunityMember{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Community{}, id).Error
	}); err != nil {
		return err
	}
	r.invalidateCommunityDelete(id, community.UserID)
	return nil
}

func (r *communityRepository) DeactivateByOwnerIDs(ownerIDs []uint) error {
	if len(ownerIDs) == 0 {
		return nil
	}
	return r.writeDB.Model(&model.Community{}).
		Where("user_id IN ?", ownerIDs).
		Update("is_active", false).Error
}

func (r *communityRepository) FindAll(page, limit int, category string) ([]model.Community, int64, error) {
	ctx := context.Background()
	key := communityListKey(category, page, limit)
	var result cachedCommunities
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return result.Communities, result.Total, nil
		}
	}

	var communities []model.Community
	var total int64
	q := r.readDB.Model(&model.Community{})
	if category != "" {
		q = q.Where("LOWER(category) = LOWER(?)", category)
	}
	q.Count(&total)
	offset := (page - 1) * limit
	if err := q.Preload("User").Order("created_at DESC").Offset(offset).Limit(limit).Find(&communities).Error; err != nil {
		return nil, 0, err
	}
	if r.cache != nil {
		result = cachedCommunities{Communities: communities, Total: total}
		_ = r.cache.Set(ctx, key, result, cache.DefaultTTL)
	}
	return communities, total, nil
}

func (r *communityRepository) FindByCreatorID(userID uint, page, limit int) ([]model.Community, int64, error) {
	ctx := context.Background()
	key := communityCreatedKey(userID, page, limit)
	var result cachedCommunities
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return result.Communities, result.Total, nil
		}
	}

	var communities []model.Community
	var total int64
	r.readDB.Model(&model.Community{}).Where("user_id = ?", userID).Count(&total)
	offset := (page - 1) * limit
	if err := r.readDB.Preload("User").Where("user_id = ?", userID).Order("created_at DESC").Offset(offset).Limit(limit).Find(&communities).Error; err != nil {
		return nil, 0, err
	}
	if r.cache != nil {
		result = cachedCommunities{Communities: communities, Total: total}
		_ = r.cache.Set(ctx, key, result, cache.DefaultTTL)
	}
	return communities, total, nil
}

func (r *communityRepository) FindJoinedByUserID(userID uint, page, limit int) ([]model.Community, int64, error) {
	ctx := context.Background()
	key := communityJoinedKey(userID, page, limit)
	var result cachedCommunities
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return result.Communities, result.Total, nil
		}
	}

	var communities []model.Community
	var total int64
	r.readDB.Model(&model.Community{}).
		Joins("JOIN community_members ON community_members.community_id = communities.community_id").
		Where("community_members.user_id = ?", userID).
		Count(&total)
	offset := (page - 1) * limit
	if err := r.readDB.Preload("User").
		Joins("JOIN community_members ON community_members.community_id = communities.community_id").
		Where("community_members.user_id = ?", userID).
		Order("community_members.joined_at DESC").
		Offset(offset).Limit(limit).
		Find(&communities).Error; err != nil {
		return nil, 0, err
	}
	if r.cache != nil {
		result = cachedCommunities{Communities: communities, Total: total}
		_ = r.cache.Set(ctx, key, result, cache.DefaultTTL)
	}
	return communities, total, nil
}

func (r *communityRepository) invalidateCommunityCreate(creatorID uint) {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.DeleteByPattern(ctx, "community:list:*")
	_ = r.cache.DeleteByPattern(ctx, fmt.Sprintf("community:created:%d:*", creatorID))
	invalidateExploreCache(ctx, r.cache)
}

func (r *communityRepository) invalidateCommunityUpdate(id, creatorID uint) {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.Delete(ctx, communityIDKey(id))
	_ = r.cache.DeleteByPattern(ctx, "community:list:*")
	_ = r.cache.DeleteByPattern(ctx, fmt.Sprintf("community:created:%d:*", creatorID))
	invalidateExploreCache(ctx, r.cache)
}

func (r *communityRepository) invalidateCommunityDelete(id, creatorID uint) {
	r.invalidateCommunityUpdate(id, creatorID)
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.DeleteByPattern(ctx, "community:joined:*")
	invalidateExploreCache(ctx, r.cache)
}
