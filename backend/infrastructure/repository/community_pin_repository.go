package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
)

type CommunityPinRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewCommunityPinRepository(writeDB, readDB *gorm.DB, c *cache.Cache) domainrepo.ICommunityPinRepository {
	return &CommunityPinRepository{writeDB: writeDB, readDB: readDB, cache: c}
}

type cachedCommunityPins struct {
	Communities []model.Community
	Total       int64
}

func communityPinListKey(userID uint, page, limit int) string {
	return fmt.Sprintf("community:pins:user:%d:page:%d:limit:%d", userID, page, limit)
}

func (r *CommunityPinRepository) Pin(pin *model.UserCommunityPin) error {
	if err := r.writeDB.Create(pin).Error; err != nil {
		return err
	}
	r.InvalidateUser(pin.UserID)
	return nil
}

func (r *CommunityPinRepository) Unpin(userID, communityID uint) error {
	if err := r.writeDB.
		Where("user_id = ? AND community_id = ?", userID, communityID).
		Delete(&model.UserCommunityPin{}).Error; err != nil {
		return err
	}
	r.InvalidateUser(userID)
	return nil
}

func (r *CommunityPinRepository) Exists(userID, communityID uint) (bool, error) {
	var count int64
	err := r.readDB.Model(&model.UserCommunityPin{}).
		Where("user_id = ? AND community_id = ?", userID, communityID).
		Count(&count).Error
	return count > 0, err
}

func (r *CommunityPinRepository) CountByUser(userID uint) (int64, error) {
	var count int64
	err := r.readDB.Model(&model.UserCommunityPin{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

func (r *CommunityPinRepository) FindMyPins(userID uint, page, limit int) ([]model.Community, int64, error) {
	ctx := context.Background()
	key := communityPinListKey(userID, page, limit)
	var result cachedCommunityPins
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return result.Communities, result.Total, nil
		}
	}

	var total int64
	r.readDB.Model(&model.Community{}).
		Joins("JOIN user_community_pins ON user_community_pins.community_id = communities.community_id").
		Where("user_community_pins.user_id = ?", userID).
		Count(&total)

	var communities []model.Community
	offset := (page - 1) * limit
	if err := r.readDB.
		Preload("User").Preload("Category").
		Joins("JOIN user_community_pins ON user_community_pins.community_id = communities.community_id").
		Where("user_community_pins.user_id = ?", userID).
		Order("user_community_pins.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&communities).Error; err != nil {
		return nil, 0, err
	}

	if r.cache != nil {
		result = cachedCommunityPins{Communities: communities, Total: total}
		if err := r.cache.Set(ctx, key, result, cache.DefaultTTL); err == nil {
			r.cache.RegisterCacheKey(ctx, communityPinsUserTag(userID), key)
		}
	}
	return communities, total, nil
}

func (r *CommunityPinRepository) InvalidateUser(userID uint) {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.InvalidateTag(ctx, communityPinsUserTag(userID))
}

func (r *CommunityPinRepository) InvalidateByCommunity(communityID uint) {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	var userIDs []uint
	if err := r.readDB.Model(&model.UserCommunityPin{}).
		Where("community_id = ?", communityID).
		Distinct("user_id").
		Pluck("user_id", &userIDs).Error; err != nil {
		return
	}
	if len(userIDs) == 0 {
		return
	}
	tags := make([]string, 0, len(userIDs))
	for _, uid := range userIDs {
		tags = append(tags, communityPinsUserTag(uid))
	}
	_ = r.cache.InvalidateTags(ctx, tags...)
}
