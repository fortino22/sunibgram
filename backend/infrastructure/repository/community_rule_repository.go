package repository

import (
	"gorm.io/gorm"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
)

type communityRuleRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewCommunityRuleRepository(writeDB, readDB *gorm.DB) domainrepo.ICommunityRuleRepository {
	return &communityRuleRepository{writeDB: writeDB, readDB: readDB}
}

func (r *communityRuleRepository) ListByCommunity(communityID uint) ([]model.CommunityRule, error) {
	var rules []model.CommunityRule
	if err := r.readDB.Where("community_id = ?", communityID).Order("order_num ASC").Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *communityRuleRepository) ListByIDsAndCommunity(communityID uint, ids []uint) ([]model.CommunityRule, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rules []model.CommunityRule
	if err := r.readDB.Where("community_id = ? AND id IN ?", communityID, ids).Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *communityRuleRepository) BulkCreateWithTx(tx *gorm.DB, rules []model.CommunityRule) error {
	if len(rules) == 0 {
		return nil
	}
	return tx.Create(&rules).Error
}

func (r *communityRuleRepository) DeleteByCommunityWithTx(tx *gorm.DB, communityID uint) error {
	return tx.Where("community_id = ?", communityID).Delete(&model.CommunityRule{}).Error
}

func (r *communityRuleRepository) DeleteByCommunityNotInWithTx(tx *gorm.DB, communityID uint, keepIDs []uint) error {
	q := tx.Where("community_id = ?", communityID)
	if len(keepIDs) > 0 {
		q = q.Where("id NOT IN ?", keepIDs)
	}
	return q.Delete(&model.CommunityRule{}).Error
}

func (r *communityRuleRepository) UpdateOrderWithTx(tx *gorm.DB, ruleID uint, orderNum int) error {
	return tx.Model(&model.CommunityRule{}).Where("id = ?", ruleID).Update("order_num", orderNum).Error
}
