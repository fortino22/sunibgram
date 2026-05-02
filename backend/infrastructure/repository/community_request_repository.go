package repository

import (
	"errors"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"gorm.io/gorm"
)

type communityRequestRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewCommunityRequestRepository(writeDB, readDB *gorm.DB) domainrepo.ICommunityRequestRepository {
	return &communityRequestRepository{writeDB: writeDB, readDB: readDB}
}

func (r *communityRequestRepository) Save(req *model.CommunityRequest) error {
	return r.writeDB.Save(req).Error
}

func (r *communityRequestRepository) FindRequest(communityID, userID uint) (*model.CommunityRequest, error) {
	var req model.CommunityRequest
	if err := r.readDB.Where("community_id = ? AND user_id = ?", communityID, userID).First(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *communityRequestRepository) UpdateStatus(communityID, userID uint, status model.CommunityRequestStatus) error {
	return r.writeDB.Model(&model.CommunityRequest{}).
		Where("community_id = ? AND user_id = ?", communityID, userID).
		UpdateColumn("status", status).Error
}

func (r *communityRequestRepository) Delete(communityID, userID uint) error {
	result := r.writeDB.Delete(&model.CommunityRequest{}, "community_id = ? AND user_id = ?", communityID, userID)
	if result.RowsAffected == 0 {
		return errors.New("request not found")
	}
	return result.Error
}

func (r *communityRequestRepository) FindPending(communityID uint, page, limit int) ([]model.CommunityRequest, int64, error) {
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
	return requests, total, nil
}
