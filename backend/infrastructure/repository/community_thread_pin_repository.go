package repository

import (
	"errors"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"gorm.io/gorm"
)

type CommunityThreadPinRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewCommunityThreadPinRepository(writeDB, readDB *gorm.DB) domainrepo.ICommunityThreadPinRepository {
	return &CommunityThreadPinRepository{writeDB: writeDB, readDB: readDB}
}

func (r *CommunityThreadPinRepository) Pin(pin *model.CommunityThreadPin) error {
	return r.writeDB.Create(pin).Error
}

func (r *CommunityThreadPinRepository) Unpin(communityID, threadID uint) error {
	return r.writeDB.Where("community_id = ? AND thread_id = ?", communityID, threadID).
		Delete(&model.CommunityThreadPin{}).Error
}

func (r *CommunityThreadPinRepository) Exists(communityID, threadID uint) (bool, error) {
	var count int64
	err := r.readDB.Model(&model.CommunityThreadPin{}).
		Where("community_id = ? AND thread_id = ?", communityID, threadID).
		Count(&count).Error
	return count > 0, err
}

func (r *CommunityThreadPinRepository) FindByCommunityID(communityID uint) (*model.CommunityThreadPin, error) {
	var pin model.CommunityThreadPin
	err := r.readDB.Where("community_id = ?", communityID).First(&pin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &pin, nil
}
