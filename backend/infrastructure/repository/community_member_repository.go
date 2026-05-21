package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
)

type communityMemberRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewCommunityMemberRepository(writeDB, readDB *gorm.DB, c *cache.Cache) domainrepo.ICommunityMemberRepository {
	return &communityMemberRepository{writeDB: writeDB, readDB: readDB, cache: c}
}

func communityMembersKey(communityID uint, page, limit int) string {
	return fmt.Sprintf("community:members:%d:page:%d:limit:%d", communityID, page, limit)
}

type cachedCommunityMembers struct {
	Members []model.CommunityMember
	Total   int64
}

func (r *communityMemberRepository) Save(member *model.CommunityMember) error {
	return r.SaveWithTx(r.writeDB, member)
}

func (r *communityMemberRepository) SaveWithTx(tx *gorm.DB, member *model.CommunityMember) error {
	member.JoinedAt = time.Now()
	if err := tx.Create(member).Error; err != nil {
		return err
	}
	r.invalidateCommunityMemberChange(member.CommunityID, member.UserID)
	return nil
}

func (r *communityMemberRepository) FindMember(communityID, userID uint) (*model.CommunityMember, error) {
	var member model.CommunityMember
	if err := r.readDB.Where("community_id = ? AND user_id = ?", communityID, userID).First(&member).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *communityMemberRepository) FindByCommunityID(communityID uint, page, limit int) ([]model.CommunityMember, int64, error) {
	ctx := context.Background()
	key := communityMembersKey(communityID, page, limit)
	var result cachedCommunityMembers
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return result.Members, result.Total, nil
		}
	}

	var members []model.CommunityMember
	var total int64
	r.readDB.Model(&model.CommunityMember{}).Where("community_id = ?", communityID).Count(&total)
	offset := (page - 1) * limit
	if err := r.readDB.Preload("User").Preload("Profile").
		Where("community_id = ?", communityID).
		Order("joined_at ASC").
		Offset(offset).Limit(limit).
		Find(&members).Error; err != nil {
		return nil, 0, err
	}
	if r.cache != nil {
		result = cachedCommunityMembers{Members: members, Total: total}
		_ = r.cache.Set(ctx, key, result, cache.DefaultTTL)
	}
	return members, total, nil
}

func (r *communityMemberRepository) FindAdmins(communityID uint) ([]model.CommunityMember, error) {
	ctx := context.Background()
	key := fmt.Sprintf("community:about:admins:%d", communityID)
	var cached []model.CommunityMember
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &cached); err == nil {
			return cached, nil
		}
	}
	var admins []model.CommunityMember
	if err := r.readDB.Preload("User").Preload("Profile").
		Where("community_id = ? AND role = ?", communityID, model.CommunityMemberRoleAdmin).
		Order("joined_at ASC").
		Find(&admins).Error; err != nil {
		return nil, err
	}
	if r.cache != nil {
		_ = r.cache.Set(ctx, key, admins, cache.DefaultTTL)
	}
	return admins, nil
}

func (r *communityMemberRepository) FindTopMembers(communityID uint, limit int) ([]model.CommunityMember, error) {
	if limit <= 0 {
		limit = 10
	}
	ctx := context.Background()
	key := fmt.Sprintf("community:about:topmembers:%d:limit:%d", communityID, limit)
	var cached []model.CommunityMember
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &cached); err == nil {
			return cached, nil
		}
	}
	var members []model.CommunityMember
	if err := r.readDB.Preload("User").Preload("Profile").
		Where("community_id = ? AND role = ?", communityID, model.CommunityMemberRoleMember).
		Order("joined_at ASC").
		Limit(limit).
		Find(&members).Error; err != nil {
		return nil, err
	}
	if r.cache != nil {
		_ = r.cache.Set(ctx, key, members, cache.DefaultTTL)
	}
	return members, nil
}

func (r *communityMemberRepository) Delete(communityID, userID uint) error {
	result := r.writeDB.Delete(&model.CommunityMember{}, "community_id = ? AND user_id = ?", communityID, userID)
	if result.RowsAffected == 0 {
		return errors.New("member not found")
	}
	if result.Error != nil {
		return result.Error
	}
	r.invalidateCommunityMemberChange(communityID, userID)
	return nil
}

func (r *communityMemberRepository) UpdateRole(communityID, userID uint, role model.CommunityMemberRole) error {
	if err := r.writeDB.Model(&model.CommunityMember{}).
		Where("community_id = ? AND user_id = ?", communityID, userID).
		UpdateColumn("role", role).Error; err != nil {
		return err
	}
	r.invalidateCommunityMemberRole(communityID)
	return nil
}

func (r *communityMemberRepository) UpdateMemberCount(communityID uint, delta int) error {
	var expr interface{}
	if delta >= 0 {
		expr = gorm.Expr("member_count + ?", delta)
	} else {
		expr = gorm.Expr("GREATEST(member_count + ?, 0)", delta)
	}
	if err := r.writeDB.Model(&model.Community{}).
		Where("community_id = ?", communityID).
		UpdateColumn("member_count", expr).Error; err != nil {
		return err
	}
	if r.cache != nil {
		ctx := context.Background()
		_ = r.cache.Delete(ctx, communityIDKey(communityID))
		_ = r.cache.DeleteByPattern(ctx, "community:joined:*")
		_ = r.cache.DeleteByPattern(ctx, "community:created:*")
		invalidateExploreCache(ctx, r.cache)
	}
	return nil
}

func (r *communityMemberRepository) invalidateCommunityMemberChange(communityID, userID uint) {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.DeleteByPattern(ctx, fmt.Sprintf("community:members:%d:*", communityID))
	_ = r.cache.DeleteByPattern(ctx, fmt.Sprintf("community:joined:%d:*", userID))
	_ = r.cache.Delete(ctx, communityIDKey(communityID))
	_ = r.cache.Delete(ctx, fmt.Sprintf("community:about:admins:%d", communityID))
	_ = r.cache.DeleteByPattern(ctx, fmt.Sprintf("community:about:topmembers:%d:*", communityID))
	_ = r.cache.DeleteByPattern(ctx, fmt.Sprintf("community:about:%d:*", communityID))
}

// Flush community:members:* manually after a backend deploy that changes what
// FindByCommunityID preloads or serializes. Use:
// redis-cli --scan --pattern "community:members:*" | xargs redis-cli del
// or, from the app cache layer:
// cache.DeleteByPattern(ctx, "community:members:*")
func (r *communityMemberRepository) invalidateCommunityMemberRole(communityID uint) {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.DeleteByPattern(ctx, fmt.Sprintf("community:members:%d:*", communityID))
	_ = r.cache.Delete(ctx, fmt.Sprintf("community:about:admins:%d", communityID))
	_ = r.cache.DeleteByPattern(ctx, fmt.Sprintf("community:about:topmembers:%d:*", communityID))
	_ = r.cache.DeleteByPattern(ctx, fmt.Sprintf("community:about:%d:*", communityID))
}
