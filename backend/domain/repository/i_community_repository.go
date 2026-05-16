package repository

import (
	"github.com/sunibgram/domain/model"
	"gorm.io/gorm"
)

type ICommunityRepository interface {
	Save(community *model.Community) error
	SaveWithTx(tx *gorm.DB, community *model.Community) error
	Transaction(fn func(tx *gorm.DB) error) error
	FindByID(id uint) (*model.Community, error)
	FindDiscover() ([]model.DiscoverCommunity, error)
	Update(community *model.Community) error
	Delete(id uint) error
	DeactivateByOwnerIDs(ownerIDs []uint) error
	ReactivateByOwnerID(ownerID uint) error
	FindByCreatorID(userID uint, page, limit int) ([]model.Community, int64, error)
	FindJoinedByUserID(userID uint, page, limit int) ([]model.Community, int64, error)
}
