package service

import (
	"context"

	communitymembermodel "github.com/sunibgram/domain/model/community_member"
)

type ICommunityMemberService interface {
	Join(ctx context.Context, communityID, userID uint) error
	Leave(ctx context.Context, communityID, userID uint) error
	GetMembers(communityID uint, page, limit int) ([]communitymembermodel.CommunityMember, int64, error)
	GetCommunityUsers(ctx context.Context, communityID, viewerID uint, role *communitymembermodel.CommunityMemberRole, query string, page, limit int) ([]CommunityUserItem, int64, error)
	RemoveMember(ctx context.Context, communityID, actorUserID, targetUserID uint) error
	UpdateRole(ctx context.Context, communityID, actorUserID, targetUserID uint, role communitymembermodel.CommunityMemberRole) error
	IsMember(communityID, userID uint) bool
}

type CommunityUserItem struct {
	UserID           uint
	Username         string
	Name             string
	AvatarObjectName string
	AvatarBucket     string
	IsPremium        bool
	Role             communitymembermodel.CommunityMemberRole
	IsFollowing      bool
}
