package repository

import (
	"gorm.io/gorm"

	"github.com/sunibgram/domain/model"
)

type ICommunityMemberRepository interface {
	Save(member *model.CommunityMember) error
	SaveWithTx(tx *gorm.DB, member *model.CommunityMember) error
	FindMember(communityID, userID uint) (*model.CommunityMember, error)
	FindByCommunityID(communityID uint, page, limit int) ([]model.CommunityMember, int64, error)
	FindCommunityUsers(communityID uint, role *model.CommunityMemberRole, query string, page, limit int) ([]model.CommunityMember, int64, error)
	FindAdmins(communityID uint) ([]model.CommunityMember, error)
	FindTopMembers(communityID uint, limit int) ([]model.CommunityMember, error)
	Delete(communityID, userID uint) error
	UpdateMemberCount(communityID uint, delta int) error
	UpdateRole(communityID, userID uint, role model.CommunityMemberRole) error
}
