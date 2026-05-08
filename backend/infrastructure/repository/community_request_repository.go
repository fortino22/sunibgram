package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
)

type communityRequestRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewCommunityRequestRepository(writeDB, readDB *gorm.DB, c *cache.Cache) domainrepo.ICommunityRequestRepository {
	return &communityRequestRepository{writeDB: writeDB, readDB: readDB, cache: c}
}

func communityRequestsKey(communityID uint, page, limit int) string {
	return fmt.Sprintf("community:requests:%d:page:%d:limit:%d", communityID, page, limit)
}

type cachedCommunityRequests struct {
	Requests []model.CommunityRequest
	Total    int64
}

func (r *communityRequestRepository) Save(req *model.CommunityRequest) error {
	if err := r.writeDB.Save(req).Error; err != nil {
		return err
	}
	r.invalidateCommunityRequests(req.CommunityID)
	return nil
}

func (r *communityRequestRepository) FindRequest(communityID, userID uint) (*model.CommunityRequest, error) {
	var req model.CommunityRequest
	if err := r.readDB.Where("community_id = ? AND user_id = ?", communityID, userID).First(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *communityRequestRepository) UpdateStatus(communityID, userID uint, status model.CommunityRequestStatus) error {
	if err := r.writeDB.Model(&model.CommunityRequest{}).
		Where("community_id = ? AND user_id = ?", communityID, userID).
		UpdateColumn("status", status).Error; err != nil {
		return err
	}
	r.invalidateCommunityRequests(communityID)
	return nil
}

func (r *communityRequestRepository) Delete(communityID, userID uint) error {
	result := r.writeDB.Delete(&model.CommunityRequest{}, "community_id = ? AND user_id = ?", communityID, userID)
	if result.RowsAffected == 0 {
		return errors.New("request not found")
	}
	if result.Error != nil {
		return result.Error
	}
	r.invalidateCommunityRequests(communityID)
	return nil
}

func (r *communityRequestRepository) FindPending(communityID uint, page, limit int) ([]model.CommunityRequest, int64, error) {
	ctx := context.Background()
	key := communityRequestsKey(communityID, page, limit)
	var result cachedCommunityRequests
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return result.Requests, result.Total, nil
		}
	}

	var requests []model.CommunityRequest
	var total int64
	r.readDB.Model(&model.CommunityRequest{}).
		Where("community_id = ? AND status = ?", communityID, model.CommunityRequestStatusPending).
		Count(&total)
	offset := (page - 1) * limit
	if err := r.readDB.Preload("User").
		Where("community_id = ? AND status = ?", communityID, model.CommunityRequestStatusPending).
		Order("created_at ASC").
		Offset(offset).Limit(limit).
		Find(&requests).Error; err != nil {
		return nil, 0, err
	}
	if r.cache != nil {
		result = cachedCommunityRequests{Requests: requests, Total: total}
		_ = r.cache.Set(ctx, key, result, cache.DefaultTTL)
	}
	return requests, total, nil
}

func (r *communityRequestRepository) invalidateCommunityRequests(communityID uint) {
	if r.cache == nil {
		return
	}
	_ = r.cache.DeleteByPattern(context.Background(), fmt.Sprintf("community:requests:%d:*", communityID))
}
