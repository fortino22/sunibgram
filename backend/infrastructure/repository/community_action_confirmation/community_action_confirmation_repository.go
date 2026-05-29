package communityactionconfirmationrepo

import (
	"gorm.io/gorm"

	communityactionconfirmationmodel "github.com/sunibgram/domain/model/community_action_confirmation"
	communityactionconfirmationrepoiface "github.com/sunibgram/domain/repository/community_action_confirmation"
)

type communityActionConfirmationRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewCommunityActionConfirmationRepository(writeDB, readDB *gorm.DB) communityactionconfirmationrepoiface.ICommunityActionConfirmationRepository {
	return &communityActionConfirmationRepository{writeDB: writeDB, readDB: readDB}
}

func (r *communityActionConfirmationRepository) Save(rec *communityactionconfirmationmodel.CommunityActionConfirmation) error {
	return r.writeDB.Create(rec).Error
}
