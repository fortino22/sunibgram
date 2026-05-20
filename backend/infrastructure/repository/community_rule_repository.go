package repository

import (
	"errors"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"gorm.io/gorm"
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

func (r *communityRuleRepository) FindByID(id uint) (*model.CommunityRule, error) {
	var rule model.CommunityRule
	if err := r.readDB.First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *communityRuleRepository) Create(rule model.CommunityRule) (*model.CommunityRule, error) {
	if err := r.writeDB.Create(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *communityRuleRepository) Update(rule model.CommunityRule) (*model.CommunityRule, error) {
	result := r.writeDB.Model(&model.CommunityRule{}).
		Where("id = ? AND community_id = ?", rule.ID, rule.CommunityID).
		UpdateColumns(map[string]any{
			"order_num":  rule.OrderNum,
			"text":       rule.Text,
			"updated_at": gorm.Expr("NOW()"),
		})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, errors.New("rule not found")
	}
	return r.FindByID(rule.ID)
}

func (r *communityRuleRepository) Delete(id uint) error {
	result := r.writeDB.Delete(&model.CommunityRule{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("rule not found")
	}
	return nil
}

func (r *communityRuleRepository) CountByCommunity(communityID uint) (int, error) {
	var count int64
	if err := r.readDB.Model(&model.CommunityRule{}).Where("community_id = ?", communityID).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}
