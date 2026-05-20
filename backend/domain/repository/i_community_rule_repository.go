package repository

import "github.com/sunibgram/domain/model"

type ICommunityRuleRepository interface {
	ListByCommunity(communityID uint) ([]model.CommunityRule, error)
	FindByID(id uint) (*model.CommunityRule, error)
	Create(rule model.CommunityRule) (*model.CommunityRule, error)
	Update(rule model.CommunityRule) (*model.CommunityRule, error)
	Delete(id uint) error
	CountByCommunity(communityID uint) (int, error)
}
