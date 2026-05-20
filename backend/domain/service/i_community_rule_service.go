package service

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type ICommunityRuleService interface {
	ListByCommunity(communityID uint) ([]model.CommunityRule, error)
	Create(ctx context.Context, communityID, actorUserID uint, orderNum int, text string) (*model.CommunityRule, error)
	Update(ctx context.Context, communityID, actorUserID, ruleID uint, orderNum int, text string) (*model.CommunityRule, error)
	Delete(ctx context.Context, communityID, actorUserID, ruleID uint) error
}
