package service

import (
	"github.com/sunibgram/domain/model"
)

type ICommunityRuleService interface {
	ListByCommunity(communityID uint) ([]model.CommunityRule, error)
}
