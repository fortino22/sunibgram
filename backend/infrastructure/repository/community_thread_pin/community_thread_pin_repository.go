package communitythreadpinrepo

import (
	"errors"

	"gorm.io/gorm"

	communitythreadpinmodel "github.com/sunibgram/domain/model/community_thread_pin"
	communitythreadpinrepoiface "github.com/sunibgram/domain/repository/community_thread_pin"
)

type CommunityThreadPinRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewCommunityThreadPinRepository(writeDB, readDB *gorm.DB) communitythreadpinrepoiface.ICommunityThreadPinRepository {
	return &CommunityThreadPinRepository{writeDB: writeDB, readDB: readDB}
}

func (r *CommunityThreadPinRepository) Pin(pin *communitythreadpinmodel.CommunityThreadPin) error {
	return r.writeDB.Create(pin).Error
}

func (r *CommunityThreadPinRepository) Unpin(communityID, threadID uint) error {
	return r.writeDB.Where("community_id = ? AND thread_id = ?", communityID, threadID).
		Delete(&communitythreadpinmodel.CommunityThreadPin{}).Error
}

func (r *CommunityThreadPinRepository) Exists(communityID, threadID uint) (bool, error) {
	var count int64
	err := r.readDB.Model(&communitythreadpinmodel.CommunityThreadPin{}).
		Where("community_id = ? AND thread_id = ?", communityID, threadID).
		Count(&count).Error
	return count > 0, err
}

func (r *CommunityThreadPinRepository) FindByCommunityID(communityID uint) (*communitythreadpinmodel.CommunityThreadPin, error) {
	var pin communitythreadpinmodel.CommunityThreadPin
	err := r.readDB.Where("community_id = ?", communityID).First(&pin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &pin, nil
}
