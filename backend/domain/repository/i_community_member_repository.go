package repository

import (
	"github.com/sunibgram/domain/model"
	"gorm.io/gorm"
)

type ICommunityMemberRepository interface {
	Save(member *model.CommunityMember) error
	SaveWithTx(tx *gorm.DB, member *model.CommunityMember) error
	FindMember(communityID, userID uint) (*model.CommunityMember, error)
	FindByCommunityID(communityID uint, page, limit int) ([]model.CommunityMember, int64, error)
	Delete(communityID, userID uint) error
	UpdateMemberCount(communityID uint, delta int) error
	UpdateRole(communityID, userID uint, role model.CommunityMemberRole) error
}
