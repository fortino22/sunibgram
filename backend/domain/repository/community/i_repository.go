package communityrepoiface

import (
	"gorm.io/gorm"

	communitymodel "github.com/sunibgram/domain/model/community"
	communitydiscovermodel "github.com/sunibgram/domain/model/community_discover"
)

type ICommunityRepository interface {
	Save(community *communitymodel.Community) error
	SaveWithTx(tx *gorm.DB, community *communitymodel.Community) error
	Transaction(fn func(tx *gorm.DB) error) error
	FindByID(id uint) (*communitymodel.Community, error)
	ExistsByName(name string, excludeID *uint) (bool, error)
	FindDiscover() ([]communitydiscovermodel.DiscoverCommunity, error)
	FindDiscoverByName(name string) ([]communitydiscovermodel.DiscoverCommunity, error)
	Update(community *communitymodel.Community) error
	UpdateWithTx(tx *gorm.DB, community *communitymodel.Community) error
	Delete(id uint) error
	DeactivateByOwnerIDs(ownerIDs []uint) error
	ReactivateByOwnerID(ownerID uint) error
	FindByCreatorID(userID uint, page, limit int) ([]communitymodel.Community, int64, error)
	FindJoinedByUserID(userID uint, page, limit int) ([]communitymodel.Community, int64, error)
}
