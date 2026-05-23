package communityrepo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	categorymodel "github.com/sunibgram/domain/model/category"
	communitymodel "github.com/sunibgram/domain/model/community"
	communitydiscovermodel "github.com/sunibgram/domain/model/community_discover"
	communitymembermodel "github.com/sunibgram/domain/model/community_member"
	communityrepoiface "github.com/sunibgram/domain/repository/community"
	"github.com/sunibgram/infrastructure/cache"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
)

type communityRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewCommunityRepository(writeDB, readDB *gorm.DB, c *cache.Cache) communityrepoiface.ICommunityRepository {
	return &communityRepository{writeDB: writeDB, readDB: readDB, cache: c}
}

func CommunityIDKey(id uint) string {
	return fmt.Sprintf("community:id:v2:%d", id)
}

func communityCreatedKey(userID uint, page, limit int) string {
	return fmt.Sprintf("community:created:%d:page:%d:limit:%d", userID, page, limit)
}

func communityJoinedKey(userID uint, page, limit int) string {
	return fmt.Sprintf("community:joined:%d:page:%d:limit:%d", userID, page, limit)
}

type cachedCommunities struct {
	Communities []communitymodel.Community
	Total       int64
}

func (r *communityRepository) Save(community *communitymodel.Community) error {
	return r.SaveWithTx(r.writeDB, community)
}

func (r *communityRepository) SaveWithTx(tx *gorm.DB, community *communitymodel.Community) error {
	if err := tx.Create(community).Error; err != nil {
		return err
	}
	r.invalidateCommunityCreate(community.UserID)
	return nil
}

func (r *communityRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.writeDB.Transaction(fn)
}

func (r *communityRepository) FindByID(id uint) (*communitymodel.Community, error) {
	ctx := context.Background()
	key := CommunityIDKey(id)
	var community communitymodel.Community
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &community); err == nil {
			return &community, nil
		}
	}
	if err := r.readDB.Preload("User").Preload("Category").First(&community, id).Error; err != nil {
		return nil, err
	}
	topMembers, err := r.findTopMembers([]uint{id})
	if err != nil {
		return nil, err
	}
	community.TopMembers = topMembers[id]
	if r.cache != nil {
		if err := r.cache.Set(ctx, key, community, cache.DefaultTTL); err == nil {
			r.cache.RegisterCacheKey(ctx, infrarepoShared.CommunityIDGlobalTag(), key)
		}
	}
	return &community, nil
}

func (r *communityRepository) ExistsByName(name string, excludeID *uint) (bool, error) {
	var count int64
	query := r.readDB.Model(&communitymodel.Community{}).Where("LOWER(name) = LOWER(?)", name)
	if excludeID != nil {
		query = query.Where("community_id <> ?", *excludeID)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *communityRepository) FindDiscover() ([]communitydiscovermodel.DiscoverCommunity, error) {
	type communityRow struct {
		CommunityID       uint
		Name              string
		BannerObjectName  string
		CategoryID        *uuid.UUID
		CategoryName      string
		CategoryCreatedAt *time.Time
		MemberCount       int
	}

	var rows []communityRow
	if err := r.readDB.
		Table("communities").
		Select(`
			communities.community_id,
			communities.name,
			COALESCE(communities.banner_object_name, '') AS banner_object_name,
			categories.id AS category_id,
			COALESCE(categories.name, '') AS category_name,
			categories.created_at AS category_created_at,
			COUNT(community_members.user_id) AS member_count
		`).
		Joins("LEFT JOIN categories ON categories.id = communities.category_id").
		Joins("LEFT JOIN community_members ON community_members.community_id = communities.community_id").
		Group("communities.community_id, communities.name, communities.banner_object_name, categories.id, categories.name, categories.created_at").
		Order("member_count DESC, communities.name ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	communities := make([]communitydiscovermodel.DiscoverCommunity, len(rows))
	communityIDs := make([]uint, 0, len(rows))
	byID := make(map[uint]int, len(rows))
	for i, row := range rows {
		var category *categorymodel.Category
		if row.CategoryID != nil {
			category = &categorymodel.Category{
				ID:        *row.CategoryID,
				Name:      row.CategoryName,
				CreatedAt: *row.CategoryCreatedAt,
			}
		}
		communities[i] = communitydiscovermodel.DiscoverCommunity{
			CommunityID:      row.CommunityID,
			Name:             row.Name,
			BannerObjectName: row.BannerObjectName,
			Category:         category,
			MemberCount:      row.MemberCount,
			TopMembers:       []communitydiscovermodel.DiscoverCommunityMember{},
		}
		communityIDs = append(communityIDs, row.CommunityID)
		byID[row.CommunityID] = i
	}
	if len(communityIDs) == 0 {
		return communities, nil
	}

	topMembers, err := r.findTopMembers(communityIDs)
	if err != nil {
		return nil, err
	}

	for communityID, members := range topMembers {
		if idx, ok := byID[communityID]; ok {
			communities[idx].TopMembers = members
		}
	}

	return communities, nil
}

func (r *communityRepository) FindDiscoverByName(name string) ([]communitydiscovermodel.DiscoverCommunity, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return r.FindDiscover()
	}

	type communityRow struct {
		CommunityID       uint
		Name              string
		BannerObjectName  string
		CategoryID        *uuid.UUID
		CategoryName      string
		CategoryCreatedAt *time.Time
		MemberCount       int
	}

	var rows []communityRow
	if err := r.readDB.
		Table("communities").
		Select(`
			communities.community_id,
			communities.name,
			COALESCE(communities.banner_object_name, '') AS banner_object_name,
			categories.id AS category_id,
			COALESCE(categories.name, '') AS category_name,
			categories.created_at AS category_created_at,
			COUNT(community_members.user_id) AS member_count
		`).
		Joins("LEFT JOIN categories ON categories.id = communities.category_id").
		Joins("LEFT JOIN community_members ON community_members.community_id = communities.community_id").
		Where("LOWER(communities.name) LIKE LOWER(?)", "%"+name+"%").
		Group("communities.community_id, communities.name, communities.banner_object_name, categories.id, categories.name, categories.created_at").
		Order("member_count DESC, communities.name ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	communities := make([]communitydiscovermodel.DiscoverCommunity, len(rows))
	communityIDs := make([]uint, 0, len(rows))
	byID := make(map[uint]int, len(rows))
	for i, row := range rows {
		var category *categorymodel.Category
		if row.CategoryID != nil {
			category = &categorymodel.Category{
				ID:        *row.CategoryID,
				Name:      row.CategoryName,
				CreatedAt: *row.CategoryCreatedAt,
			}
		}
		communities[i] = communitydiscovermodel.DiscoverCommunity{
			CommunityID:      row.CommunityID,
			Name:             row.Name,
			BannerObjectName: row.BannerObjectName,
			Category:         category,
			MemberCount:      row.MemberCount,
			TopMembers:       []communitydiscovermodel.DiscoverCommunityMember{},
		}
		communityIDs = append(communityIDs, row.CommunityID)
		byID[row.CommunityID] = i
	}
	if len(communityIDs) == 0 {
		return communities, nil
	}

	topMembers, err := r.findTopMembers(communityIDs)
	if err != nil {
		return nil, err
	}

	for communityID, members := range topMembers {
		if idx, ok := byID[communityID]; ok {
			communities[idx].TopMembers = members
		}
	}

	return communities, nil
}

func (r *communityRepository) findTopMembers(communityIDs []uint) (map[uint][]communitydiscovermodel.DiscoverCommunityMember, error) {
	topMembers := make(map[uint][]communitydiscovermodel.DiscoverCommunityMember, len(communityIDs))
	if len(communityIDs) == 0 {
		return topMembers, nil
	}

	type memberRow struct {
		CommunityID      uint
		UserID           uint
		AvatarObjectName string
	}

	var memberRows []memberRow
	if err := r.readDB.
		Raw(`
			SELECT community_id, user_id, avatar_object_name
			FROM (
				SELECT
					community_members.community_id,
					community_members.user_id,
					COALESCE(user_profiles.avatar_object_name, '') AS avatar_object_name,
					ROW_NUMBER() OVER (
						PARTITION BY community_members.community_id
						ORDER BY COALESCE(user_stats.follower_count, 0) DESC,
							community_members.joined_at ASC,
							community_members.user_id ASC
					) AS member_rank
				FROM community_members
				JOIN users ON users.user_id = community_members.user_id
				LEFT JOIN user_profiles ON user_profiles.user_id = users.user_id
				LEFT JOIN user_stats ON user_stats.user_id = users.user_id
				WHERE community_members.community_id IN ?
			) ranked_members
			WHERE member_rank <= 5
			ORDER BY community_id ASC, member_rank ASC
		`, communityIDs).
		Scan(&memberRows).Error; err != nil {
		return nil, err
	}

	for _, row := range memberRows {
		topMembers[row.CommunityID] = append(topMembers[row.CommunityID], communitydiscovermodel.DiscoverCommunityMember{
			UserID:           row.UserID,
			AvatarObjectName: row.AvatarObjectName,
		})
	}

	return topMembers, nil
}

func (r *communityRepository) Update(community *communitymodel.Community) error {
	return r.UpdateWithTx(r.writeDB, community)
}

func (r *communityRepository) UpdateWithTx(tx *gorm.DB, community *communitymodel.Community) error {
	var categoryID any
	if community.CategoryID != nil {
		categoryID = *community.CategoryID
	}
	if err := tx.Model(&communitymodel.Community{}).Where("community_id = ?", community.CommunityID).UpdateColumns(map[string]any{
		"name":               community.Name,
		"description":        community.Description,
		"category_id":        categoryID,
		"visibility":         community.Visibility,
		"banner_object_name": community.BannerObjectName,
		"updated_at":         gorm.Expr("NOW()"),
	}).Error; err != nil {
		return err
	}
	r.invalidateCommunityUpdate(community.CommunityID, community.UserID)
	return nil
}

func (r *communityRepository) Delete(id uint) error {
	var community communitymodel.Community
	_ = r.writeDB.Select("community_id", "user_id").First(&community, id).Error
	if err := r.writeDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("community_id = ?", id).Delete(&communitymembermodel.CommunityMember{}).Error; err != nil {
			return err
		}
		return tx.Delete(&communitymodel.Community{}, id).Error
	}); err != nil {
		return err
	}
	r.invalidateCommunityDelete(id, community.UserID)
	return nil
}

func (r *communityRepository) DeactivateByOwnerIDs(ownerIDs []uint) error {
	if len(ownerIDs) == 0 {
		return nil
	}
	var communities []communitymodel.Community
	_ = r.writeDB.Select("community_id").Where("user_id IN ?", ownerIDs).Find(&communities).Error
	if err := r.writeDB.Model(&communitymodel.Community{}).
		Where("user_id IN ?", ownerIDs).
		Update("is_active", false).Error; err != nil {
		return err
	}
	if r.cache != nil {
		ctx := context.Background()
		for _, c := range communities {
			_ = r.cache.Delete(ctx, CommunityIDKey(c.CommunityID))
			r.invalidateCommunityThreadCaches(ctx, c.CommunityID)
		}
		infrarepoShared.InvalidateExploreCache(ctx, r.cache)
	}
	return nil
}

func (r *communityRepository) ReactivateByOwnerID(ownerID uint) error {
	var communities []communitymodel.Community
	_ = r.writeDB.Select("community_id").Where("user_id = ?", ownerID).Find(&communities).Error
	if err := r.writeDB.Model(&communitymodel.Community{}).
		Where("user_id = ?", ownerID).
		Update("is_active", true).Error; err != nil {
		return err
	}
	if r.cache != nil {
		ctx := context.Background()
		for _, c := range communities {
			_ = r.cache.Delete(ctx, CommunityIDKey(c.CommunityID))
			r.invalidateCommunityThreadCaches(ctx, c.CommunityID)
		}
		infrarepoShared.InvalidateExploreCache(ctx, r.cache)
	}
	return nil
}

func (r *communityRepository) FindByCreatorID(userID uint, page, limit int) ([]communitymodel.Community, int64, error) {
	ctx := context.Background()
	key := communityCreatedKey(userID, page, limit)
	var result cachedCommunities
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return result.Communities, result.Total, nil
		}
	}

	var communities []communitymodel.Community
	var total int64
	r.readDB.Model(&communitymodel.Community{}).Where("user_id = ?", userID).Count(&total)
	offset := (page - 1) * limit
	if err := r.readDB.Preload("User").Preload("Category").Where("user_id = ?", userID).Order("created_at DESC").Offset(offset).Limit(limit).Find(&communities).Error; err != nil {
		return nil, 0, err
	}
	if r.cache != nil {
		result = cachedCommunities{Communities: communities, Total: total}
		if err := r.cache.Set(ctx, key, result, cache.DefaultTTL); err == nil {
			r.cache.RegisterCacheKey(ctx, infrarepoShared.CommunityListsByUserTag(userID), key)
			r.cache.RegisterCacheKey(ctx, infrarepoShared.CommunityListsGlobalTag(), key)
		}
	}
	return communities, total, nil
}

func (r *communityRepository) FindJoinedByUserID(userID uint, page, limit int) ([]communitymodel.Community, int64, error) {
	ctx := context.Background()
	key := communityJoinedKey(userID, page, limit)
	var result cachedCommunities
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &result); err == nil {
			return result.Communities, result.Total, nil
		}
	}

	var communities []communitymodel.Community
	var total int64
	r.readDB.Model(&communitymodel.Community{}).
		Joins("JOIN community_members ON community_members.community_id = communities.community_id").
		Where("community_members.user_id = ?", userID).
		Count(&total)
	offset := (page - 1) * limit
	if err := r.readDB.Preload("User").Preload("Category").
		Joins("JOIN community_members ON community_members.community_id = communities.community_id").
		Where("community_members.user_id = ?", userID).
		Order("community_members.joined_at DESC").
		Offset(offset).Limit(limit).
		Find(&communities).Error; err != nil {
		return nil, 0, err
	}
	if r.cache != nil {
		result = cachedCommunities{Communities: communities, Total: total}
		if err := r.cache.Set(ctx, key, result, cache.DefaultTTL); err == nil {
			r.cache.RegisterCacheKey(ctx, infrarepoShared.CommunityListsByUserTag(userID), key)
			r.cache.RegisterCacheKey(ctx, infrarepoShared.CommunityListsGlobalTag(), key)
		}
	}
	return communities, total, nil
}

func (r *communityRepository) invalidateCommunityCreate(creatorID uint) {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.InvalidateTag(ctx, infrarepoShared.CommunityListsByUserTag(creatorID))
	infrarepoShared.InvalidateExploreCache(ctx, r.cache)
}

func (r *communityRepository) invalidateCommunityUpdate(id, creatorID uint) {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.Delete(ctx, CommunityIDKey(id), fmt.Sprintf("community:about:admins:%d", id))
	_ = r.cache.InvalidateTags(ctx,
		infrarepoShared.CommunityListsByUserTag(creatorID),
		infrarepoShared.CommunityMembersTag(id),
	)
	r.invalidateCommunityThreadCaches(ctx, id)
	infrarepoShared.InvalidateExploreCache(ctx, r.cache)
}

func (r *communityRepository) invalidateCommunityDelete(id, creatorID uint) {
	r.invalidateCommunityUpdate(id, creatorID)
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.InvalidateTag(ctx, infrarepoShared.CommunityListsGlobalTag())
	infrarepoShared.InvalidateExploreCache(ctx, r.cache)
}

func (r *communityRepository) invalidateCommunityThreadCaches(ctx context.Context, communityID uint) {
	_ = r.cache.Delete(ctx, "feed:trending:global")
	// Per-family global tags + per-community tag — replaces the migrated
	// patterns. Remaining SCAN patterns (thread:id:*, thread:userfeed/usermedia:*,
	// bookmarks/likes:user:*, feed:foryou:*) have no global tag yet — every
	// Set site for those families would have to register, costing one extra
	// SADD per write. Community update/delete is rare enough that pattern
	// SCAN is acceptable.
	_ = r.cache.InvalidateTags(ctx,
		infrarepoShared.ThreadRootsGlobalTag(),
		infrarepoShared.ThreadChainGlobalTag(),
		infrarepoShared.ThreadCommunityTag(communityID),
		infrarepoShared.CommunityThreadsGlobalTag(),
		infrarepoShared.ThreadAllCommunitiesGlobalTag(),
		infrarepoShared.ThreadMyCommunitiesGlobalTag(),
		infrarepoShared.ThreadFollowingGlobalTag(),
		infrarepoShared.EventThreadsGlobalTag(),
	)
	_ = r.cache.DeleteByPatterns(ctx,
		"thread:id:*",
		"thread:userfeed:*",
		"thread:usermedia:*",
		"feed:foryou:*",
		"bookmarks:user:*",
		"likes:user:*",
	)
}
