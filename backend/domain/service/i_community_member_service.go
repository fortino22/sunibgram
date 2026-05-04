package service

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type ICommunityMemberService interface {
	Join(ctx context.Context, communityID, userID uint) error
	Leave(ctx context.Context, communityID, userID uint) error
	GetMembers(communityID uint, page, limit int) ([]model.CommunityMember, int64, error)
	RemoveMember(ctx context.Context, communityID, actorUserID, targetUserID uint) error
	UpdateRole(ctx context.Context, communityID, actorUserID, targetUserID uint, role model.CommunityMemberRole) error
}
