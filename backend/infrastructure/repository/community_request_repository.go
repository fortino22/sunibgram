package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	communityrequestmodel "github.com/sunibgram/domain/model/community_request"
	communityrequestrepoiface "github.com/sunibgram/domain/repository/community_request"
	"github.com/sunibgram/infrastructure/cache"
)

type communityRequestRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewCommunityRequestRepository(writeDB, readDB *gorm.DB, c *cache.Cache) communityrequestrepoiface.ICommunityRequestRepository {
	return &communityRequestRepository{writeDB: writeDB, readDB: readDB, cache: c}
}

func communityRequestsKey(communityID uint, page, limit int) string {
	return fmt.Sprintf("community:requests:%d:page:%d:limit:%d", communityID, page, limit)
}

type cachedCommunityRequests struct {
	Requests []communityrequestmodel.CommunityRequest
	Total    int64
}

func (r *communityRequestRepository) Save(req *communityrequestmodel.CommunityRequest) error {
	if err := r.writeDB.Save(req).Error; err != nil {
		return err
	}
	r.invalidateCommunityRequests(req.CommunityID)
	return nil
}

func (r *communityRequestRepository) FindRequest(communityID, userID uint) (*communityrequestmodel.CommunityRequest, error) {
	var req communityrequestmodel.CommunityRequest
	if err := r.readDB.Where("community_id = ? AND user_id = ?", communityID, userID).First(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *communityRequestRepository) UpdateStatus(communityID, userID uint, status communityrequestmodel.CommunityRequestStatus) error {
	if err := r.writeDB.Model(&communityrequestmodel.CommunityRequest{}).
		Where("community_id = ? AND user_id = ?", communityID, userID).
		UpdateColumn("status", status).Error; err != nil {
		return err
	}
	r.invalidateCommunityRequests(communityID)
	return nil
}

func (r *communityRequestRepository) Delete(communityID, userID uint) error {
	result := r.writeDB.Delete(&communityrequestmodel.CommunityRequest{}, "community_id = ? AND user_id = ?", communityID, userID)
	if result.RowsAffected == 0 {
		return errors.New("request not found")
	}
	if result.Error != nil {
		return result.Error
	}
	r.invalidateCommunityRequests(communityID)
	return nil
}

func (r *communityRequestRepository) FindPending(communityID uint, page, limit int) ([]communityrequestmodel.CommunityRequest, int64, error) {
	ctx := context.Background()
	key := communityRequestsKey(communityID, page, limit)
	var result cachedCommunityRequests
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return result.Requests, result.Total, nil
		}
	}

	var requests []communityrequestmodel.CommunityRequest
	var total int64
	r.readDB.Model(&communityrequestmodel.CommunityRequest{}).
		Where("community_id = ? AND status = ?", communityID, communityrequestmodel.CommunityRequestStatusPending).
		Count(&total)
	offset := (page - 1) * limit
	if err := r.readDB.Preload("User").
		Where("community_id = ? AND status = ?", communityID, communityrequestmodel.CommunityRequestStatusPending).
		Order("created_at ASC").
		Offset(offset).Limit(limit).
		Find(&requests).Error; err != nil {
		return nil, 0, err
	}
	if r.cache != nil {
		result = cachedCommunityRequests{Requests: requests, Total: total}
		if err := r.cache.Set(ctx, key, result, cache.DefaultTTL); err == nil {
			r.cache.RegisterCacheKey(ctx, communityRequestsTag(communityID), key)
			r.cache.RegisterCacheKey(ctx, communityRequestsGlobalTag(), key)
		}
	}
	return requests, total, nil
}

func (r *communityRequestRepository) invalidateCommunityRequests(communityID uint) {
	if r.cache == nil {
		return
	}
	_ = r.cache.InvalidateTag(context.Background(), communityRequestsTag(communityID))
}
