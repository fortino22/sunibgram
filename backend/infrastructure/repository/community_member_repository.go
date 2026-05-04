package repository

import (
	"errors"
	"time"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"gorm.io/gorm"
)

type communityMemberRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewCommunityMemberRepository(writeDB, readDB *gorm.DB) domainrepo.ICommunityMemberRepository {
	return &communityMemberRepository{writeDB: writeDB, readDB: readDB}
}

func (r *communityMemberRepository) Save(member *model.CommunityMember) error {
	member.JoinedAt = time.Now()
	return r.writeDB.Create(member).Error
}

func (r *communityMemberRepository) FindMember(communityID, userID uint) (*model.CommunityMember, error) {
	var member model.CommunityMember
	if err := r.readDB.Where("community_id = ? AND user_id = ?", communityID, userID).First(&member).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *communityMemberRepository) FindByCommunityID(communityID uint, page, limit int) ([]model.CommunityMember, int64, error) {
	var members []model.CommunityMember
	var total int64
	r.readDB.Model(&model.CommunityMember{}).Where("community_id = ?", communityID).Count(&total)
	offset := (page - 1) * limit
	if err := r.readDB.Preload("User").
		Where("community_id = ?", communityID).
		Order("joined_at ASC").
		Offset(offset).Limit(limit).
		Find(&members).Error; err != nil {
		return nil, 0, err
	}
	return members, total, nil
}

func (r *communityMemberRepository) Delete(communityID, userID uint) error {
	result := r.writeDB.Delete(&model.CommunityMember{}, "community_id = ? AND user_id = ?", communityID, userID)
	if result.RowsAffected == 0 {
		return errors.New("member not found")
	}
	return result.Error
}

func (r *communityMemberRepository) UpdateRole(communityID, userID uint, role model.CommunityMemberRole) error {
	return r.writeDB.Model(&model.CommunityMember{}).
		Where("community_id = ? AND user_id = ?", communityID, userID).
		UpdateColumn("role", role).Error
}

func (r *communityMemberRepository) UpdateMemberCount(communityID uint, delta int) error {
	var expr interface{}
	if delta >= 0 {
		expr = gorm.Expr("member_count + ?", delta)
	} else {
		expr = gorm.Expr("GREATEST(member_count + ?, 0)", delta)
	}
	return r.writeDB.Model(&model.Community{}).
		Where("community_id = ?", communityID).
		UpdateColumn("member_count", expr).Error
}
