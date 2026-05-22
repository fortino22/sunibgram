package service

import (
	communityrulemodel "github.com/sunibgram/domain/model/community_rule"
)

type ICommunityRuleService interface {
	ListByCommunity(communityID uint) ([]communityrulemodel.CommunityRule, error)
}
