package communitymemberrepoiface

import (
	"gorm.io/gorm"

	communitymembermodel "github.com/sunibgram/domain/model/community_member"
)

type ICommunityMemberRepository interface {
	Save(member *communitymembermodel.CommunityMember) error
	SaveWithTx(tx *gorm.DB, member *communitymembermodel.CommunityMember) error
	FindMember(communityID, userID uint) (*communitymembermodel.CommunityMember, error)
	FindByCommunityID(communityID uint, page, limit int) ([]communitymembermodel.CommunityMember, int64, error)
	FindCommunityUsers(communityID uint, role *communitymembermodel.CommunityMemberRole, query string, page, limit int) ([]communitymembermodel.CommunityMember, int64, error)
	FindAdmins(communityID uint) ([]communitymembermodel.CommunityMember, error)
	FindTopMembers(communityID uint, limit int) ([]communitymembermodel.CommunityMember, error)
	Delete(communityID, userID uint) error
	UpdateMemberCount(communityID uint, delta int) error
	UpdateRole(communityID, userID uint, role communitymembermodel.CommunityMemberRole) error
	TransferAdmin(communityID, fromUserID, toUserID uint) error
	CountAdmins(communityID uint) (int64, error)
}
