package repository

import (
	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"gorm.io/gorm"
)

type communityRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewCommunityRepository(writeDB, readDB *gorm.DB) domainrepo.ICommunityRepository {
	return &communityRepository{writeDB: writeDB, readDB: readDB}
}

func (r *communityRepository) Save(community *model.Community) error {
	return r.SaveWithTx(r.writeDB, community)
}

func (r *communityRepository) SaveWithTx(tx *gorm.DB, community *model.Community) error {
	return tx.Create(community).Error
}

func (r *communityRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.writeDB.Transaction(fn)
}

func (r *communityRepository) FindByID(id uint) (*model.Community, error) {
	var community model.Community
	if err := r.readDB.Preload("User").First(&community, id).Error; err != nil {
		return nil, err
	}
	return &community, nil
}

func (r *communityRepository) Update(community *model.Community) error {
	return r.writeDB.Model(community).UpdateColumns(map[string]any{
		"name":               community.Name,
		"description":        community.Description,
		"rules":              community.Rules,
		"category":           community.Category,
		"visibility":         community.Visibility,
		"avatar_object_name": community.AvatarObjectName,
		"avatar_bucket":      community.AvatarBucket,
		"banner_object_name": community.BannerObjectName,
		"banner_bucket":      community.BannerBucket,
		"updated_at":         gorm.Expr("NOW()"),
	}).Error
}

func (r *communityRepository) Delete(id uint) error {
	return r.writeDB.Delete(&model.Community{}, id).Error
}

func (r *communityRepository) FindAll(page, limit int, category string) ([]model.Community, int64, error) {
	var communities []model.Community
	var total int64
	q := r.readDB.Model(&model.Community{})
	if category != "" {
		q = q.Where("LOWER(category) = LOWER(?)", category)
	}
	q.Count(&total)
	offset := (page - 1) * limit
	if err := q.Preload("User").Order("created_at DESC").Offset(offset).Limit(limit).Find(&communities).Error; err != nil {
		return nil, 0, err
	}
	return communities, total, nil
}

func (r *communityRepository) FindByCreatorID(userID uint, page, limit int) ([]model.Community, int64, error) {
	var communities []model.Community
	var total int64
	r.readDB.Model(&model.Community{}).Where("user_id = ?", userID).Count(&total)
	offset := (page - 1) * limit
	if err := r.readDB.Preload("User").Where("user_id = ?", userID).Order("created_at DESC").Offset(offset).Limit(limit).Find(&communities).Error; err != nil {
		return nil, 0, err
	}
	return communities, total, nil
}

func (r *communityRepository) FindJoinedByUserID(userID uint, page, limit int) ([]model.Community, int64, error) {
	var communities []model.Community
	var total int64
	r.readDB.Model(&model.Community{}).
		Joins("JOIN community_members ON community_members.community_id = communities.community_id").
		Where("community_members.user_id = ?", userID).
		Count(&total)
	offset := (page - 1) * limit
	if err := r.readDB.Preload("User").
		Joins("JOIN community_members ON community_members.community_id = communities.community_id").
		Where("community_members.user_id = ?", userID).
		Order("community_members.joined_at DESC").
		Offset(offset).Limit(limit).
		Find(&communities).Error; err != nil {
		return nil, 0, err
	}
	return communities, total, nil
}
