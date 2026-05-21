package service

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type ICommunityMemberService interface {
	Join(ctx context.Context, communityID, userID uint) error
	Leave(ctx context.Context, communityID, userID uint) error
	GetMembers(communityID uint, page, limit int) ([]model.CommunityMember, int64, error)
	GetCommunityUsers(ctx context.Context, communityID, viewerID uint, role *model.CommunityMemberRole, query string, page, limit int) ([]CommunityUserItem, int64, error)
	RemoveMember(ctx context.Context, communityID, actorUserID, targetUserID uint) error
	UpdateRole(ctx context.Context, communityID, actorUserID, targetUserID uint, role model.CommunityMemberRole) error
	IsMember(communityID, userID uint) bool
}

type CommunityUserItem struct {
	UserID      uint
	Username    string
	Name        string
	IsPremium   bool
	Role        model.CommunityMemberRole
	IsFollowing bool
}
