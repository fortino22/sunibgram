package repository

import (
	"github.com/sunibgram/domain/model"
	"gorm.io/gorm"
)

type ICommunityRuleRepository interface {
	ListByCommunity(communityID uint) ([]model.CommunityRule, error)
	ListByIDsAndCommunity(communityID uint, ids []uint) ([]model.CommunityRule, error)
	BulkCreateWithTx(tx *gorm.DB, rules []model.CommunityRule) error
	DeleteByCommunityWithTx(tx *gorm.DB, communityID uint) error
	DeleteByCommunityNotInWithTx(tx *gorm.DB, communityID uint, keepIDs []uint) error
	UpdateOrderWithTx(tx *gorm.DB, ruleID uint, orderNum int) error
}
