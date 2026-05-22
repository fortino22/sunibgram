package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	communitymodel "github.com/sunibgram/domain/model/community"
	communitymembermodel "github.com/sunibgram/domain/model/community_member"
	communitymemberrepoiface "github.com/sunibgram/domain/repository/community_member"
	"github.com/sunibgram/infrastructure/cache"
)

type communityMemberRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewCommunityMemberRepository(writeDB, readDB *gorm.DB, c *cache.Cache) communitymemberrepoiface.ICommunityMemberRepository {
	return &communityMemberRepository{writeDB: writeDB, readDB: readDB, cache: c}
}

func communityMembersKey(communityID uint, page, limit int) string {
	return fmt.Sprintf("community:members:%d:page:%d:limit:%d", communityID, page, limit)
}

type cachedCommunityMembers struct {
	Members []communitymembermodel.CommunityMember
	Total   int64
}

func (r *communityMemberRepository) Save(member *communitymembermodel.CommunityMember) error {
	return r.SaveWithTx(r.writeDB, member)
}

func (r *communityMemberRepository) SaveWithTx(tx *gorm.DB, member *communitymembermodel.CommunityMember) error {
	member.JoinedAt = time.Now()
	if err := tx.Create(member).Error; err != nil {
		return err
	}
	r.invalidateCommunityMemberChange(member.CommunityID, member.UserID)
	return nil
}

func (r *communityMemberRepository) FindMember(communityID, userID uint) (*communitymembermodel.CommunityMember, error) {
	var member communitymembermodel.CommunityMember
	if err := r.readDB.Where("community_id = ? AND user_id = ?", communityID, userID).First(&member).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *communityMemberRepository) FindByCommunityID(communityID uint, page, limit int) ([]communitymembermodel.CommunityMember, int64, error) {
	ctx := context.Background()
	key := communityMembersKey(communityID, page, limit)
	var result cachedCommunityMembers
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return result.Members, result.Total, nil
		}
	}

	var members []communitymembermodel.CommunityMember
	var total int64
	r.readDB.Model(&communitymembermodel.CommunityMember{}).Where("community_id = ?", communityID).Count(&total)
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
		if err := r.cache.Set(ctx, key, result, cache.DefaultTTL); err == nil {
			r.cache.RegisterCacheKey(ctx, communityMembersTag(communityID), key)
			r.cache.RegisterCacheKey(ctx, communityMembersGlobalTag(), key)
		}
	}
	return members, total, nil
}

func communityUsersKey(communityID uint, role *communitymembermodel.CommunityMemberRole, query string, page, limit int) string {
	r := "ALL"
	if role != nil {
		r = string(*role)
	}
	return fmt.Sprintf("community:users:%d:role:%s:q:%s:p:%d:l:%d", communityID, r, strings.ToLower(query), page, limit)
}

type cachedCommunityUsers struct {
	Members []communitymembermodel.CommunityMember
	Total   int64
}

func (r *communityMemberRepository) FindCommunityUsers(communityID uint, role *communitymembermodel.CommunityMemberRole, query string, page, limit int) ([]communitymembermodel.CommunityMember, int64, error) {
	ctx := context.Background()
	query = strings.TrimSpace(query)
	key := communityUsersKey(communityID, role, query, page, limit)
	var result cachedCommunityUsers
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return result.Members, result.Total, nil
		}
	}

	base := r.readDB.Model(&communitymembermodel.CommunityMember{}).
		Joins("JOIN users ON users.user_id = community_members.user_id").
		Where("community_members.community_id = ?", communityID)
	if role != nil {
		base = base.Where("community_members.role = ?", *role)
	}
	if query != "" {
		like := "%" + query + "%"
		base = base.Where("users.user_name ILIKE ? OR users.name ILIKE ?", like, like)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var members []communitymembermodel.CommunityMember
	offset := (page - 1) * limit
	q := r.readDB.Preload("User").Preload("Profile").
		Joins("JOIN users ON users.user_id = community_members.user_id").
		Where("community_members.community_id = ?", communityID)
	if role != nil {
		q = q.Where("community_members.role = ?", *role)
	}
	if query != "" {
		like := "%" + query + "%"
		q = q.Where("users.user_name ILIKE ? OR users.name ILIKE ?", like, like)
	}
	if err := q.Order("community_members.joined_at ASC").
		Offset(offset).Limit(limit).
		Find(&members).Error; err != nil {
		return nil, 0, err
	}

	if r.cache != nil {
		result = cachedCommunityUsers{Members: members, Total: total}
		if err := r.cache.Set(ctx, key, result, cache.DefaultTTL); err == nil {
			r.cache.RegisterCacheKey(ctx, communityMembersTag(communityID), key)
			r.cache.RegisterCacheKey(ctx, communityMembersGlobalTag(), key)
		}
	}
	return members, total, nil
}

func (r *communityMemberRepository) FindAdmins(communityID uint) ([]communitymembermodel.CommunityMember, error) {
	ctx := context.Background()
	key := fmt.Sprintf("community:about:admins:%d", communityID)
	var cached []communitymembermodel.CommunityMember
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &cached); err == nil {
			return cached, nil
		}
	}
	var admins []communitymembermodel.CommunityMember
	if err := r.readDB.Preload("User").Preload("Profile").
		Where("community_id = ? AND role = ?", communityID, communitymembermodel.CommunityMemberRoleAdmin).
		Order("joined_at ASC").
		Find(&admins).Error; err != nil {
		return nil, err
	}
	if r.cache != nil {
		_ = r.cache.Set(ctx, key, admins, cache.DefaultTTL)
	}
	return admins, nil
}

func (r *communityMemberRepository) FindTopMembers(communityID uint, limit int) ([]communitymembermodel.CommunityMember, error) {
	if limit <= 0 {
		limit = 10
	}
	ctx := context.Background()
	key := fmt.Sprintf("community:about:topmembers:%d:limit:%d", communityID, limit)
	var cached []communitymembermodel.CommunityMember
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &cached); err == nil {
			return cached, nil
		}
	}
	var members []communitymembermodel.CommunityMember
	if err := r.readDB.Preload("User").Preload("Profile").
		Where("community_id = ? AND role = ?", communityID, communitymembermodel.CommunityMemberRoleMember).
		Order("joined_at ASC").
		Limit(limit).
		Find(&members).Error; err != nil {
		return nil, err
	}
	if r.cache != nil {
		if err := r.cache.Set(ctx, key, members, cache.DefaultTTL); err == nil {
			r.cache.RegisterCacheKey(ctx, communityMembersTag(communityID), key)
			r.cache.RegisterCacheKey(ctx, communityMembersGlobalTag(), key)
		}
	}
	return members, nil
}

func (r *communityMemberRepository) Delete(communityID, userID uint) error {
	result := r.writeDB.Delete(&communitymembermodel.CommunityMember{}, "community_id = ? AND user_id = ?", communityID, userID)
	if result.RowsAffected == 0 {
		return errors.New("member not found")
	}
	if result.Error != nil {
		return result.Error
	}
	r.invalidateCommunityMemberChange(communityID, userID)
	return nil
}

func (r *communityMemberRepository) UpdateRole(communityID, userID uint, role communitymembermodel.CommunityMemberRole) error {
	if err := r.writeDB.Model(&communitymembermodel.CommunityMember{}).
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
	if err := r.writeDB.Model(&communitymodel.Community{}).
		Where("community_id = ?", communityID).
		UpdateColumn("member_count", expr).Error; err != nil {
		return err
	}
	if r.cache != nil {
		ctx := context.Background()
		_ = r.cache.Delete(ctx, communityIDKey(communityID))
		// communityListsGlobalTag covers community:joined/created:* across
		// every user — registered at FindByCreatorID/FindJoinedByUserID Set
		// sites.
		_ = r.cache.InvalidateTag(ctx, communityListsGlobalTag())
		invalidateExploreCache(ctx, r.cache)
	}
	return nil
}

func (r *communityMemberRepository) invalidateCommunityMemberChange(communityID, userID uint) {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.Delete(ctx, communityIDKey(communityID), fmt.Sprintf("community:about:admins:%d", communityID))
	// Tag-based — covers every legacy pattern:
	//   communityMembersTag    → community:members/users/about/about:topmembers:<cid>:*
	//   communityListsByUserTag → community:joined/created:<uid>:*
	//   userForYouTag          → feed:foryou:user:<uid>:*
	//   userFeedsTag           → thread:userfeed/usermedia/following:<uid>:*
	//   viewerThreadFeedsTag   → thread:userfeed/usermedia:*:req:<uid>,
	//                            thread:my-communities/all-communities:user:<uid>:*,
	//                            community:threads:*:user:<uid>:*,
	//                            event:threads:*:viewer:<uid>:*
	_ = r.cache.InvalidateTags(ctx,
		communityMembersTag(communityID),
		communityListsByUserTag(userID),
		userForYouTag(userID),
		userFeedsTag(userID),
		viewerThreadFeedsTag(userID),
	)
	invalidateExploreCache(ctx, r.cache)
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
	_ = r.cache.Delete(ctx, fmt.Sprintf("community:about:admins:%d", communityID))
	_ = r.cache.InvalidateTag(ctx, communityMembersTag(communityID))
}
