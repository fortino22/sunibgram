package communityrulerepoiface

import (
	"gorm.io/gorm"

	communityrulemodel "github.com/sunibgram/domain/model/community_rule"
)

type ICommunityRuleRepository interface {
	ListByCommunity(communityID uint) ([]communityrulemodel.CommunityRule, error)
	ListByIDsAndCommunity(communityID uint, ids []uint) ([]communityrulemodel.CommunityRule, error)
	BulkCreateWithTx(tx *gorm.DB, rules []communityrulemodel.CommunityRule) error
	DeleteByCommunityWithTx(tx *gorm.DB, communityID uint) error
	DeleteByCommunityNotInWithTx(tx *gorm.DB, communityID uint, keepIDs []uint) error
	UpdateOrderWithTx(tx *gorm.DB, ruleID uint, orderNum int) error
}
