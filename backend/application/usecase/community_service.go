package usecase

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/google/uuid"
	"github.com/sunibgram/application/helper"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
	"gorm.io/gorm"
)

type CommunityService struct {
	communityRepo  repository.ICommunityRepository
	memberRepo     repository.ICommunityMemberRepository
	userRepo       repository.IUserRepository
	categoryRepo   repository.ICategoryRepository
	ruleRepo       repository.ICommunityRuleRepository
	cacheRepo      repository.ICacheRepository
	storageUsecase *StorageUsecase
}

func NewCommunityService(
	communityRepo repository.ICommunityRepository,
	memberRepo repository.ICommunityMemberRepository,
	userRepo repository.IUserRepository,
	categoryRepo repository.ICategoryRepository,
	ruleRepo repository.ICommunityRuleRepository,
	cacheRepo repository.ICacheRepository,
	storageUsecase *StorageUsecase,
) service.ICommunityService {
	return &CommunityService{
		communityRepo:  communityRepo,
		memberRepo:     memberRepo,
		userRepo:       userRepo,
		categoryRepo:   categoryRepo,
		ruleRepo:       ruleRepo,
		cacheRepo:      cacheRepo,
		storageUsecase: storageUsecase,
	}
}

func (s *CommunityService) Create(ctx context.Context, userID uint, name, description string, rules []string, categoryID *uuid.UUID, visibility model.CommunityVisibility, banner *multipart.FileHeader) (*model.Community, []model.CommunityRule, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil, errors.New("name is required")
	}
	if err := s.ensureCommunityNameAvailable(name, nil); err != nil {
		return nil, nil, err
	}
	if visibility != model.CommunityVisibilityPublic && visibility != model.CommunityVisibilityPrivate {
		visibility = model.CommunityVisibilityPublic
	}
	if err := s.validateCategoryID(categoryID); err != nil {
		return nil, nil, err
	}
	cleanedRules, err := sanitizeCommunityRules(rules)
	if err != nil {
		return nil, nil, err
	}

	community := &model.Community{
		UserID:      userID,
		Name:        name,
		Description: description,
		CategoryID:  categoryID,
		Visibility:  visibility,
		MemberCount: 1,
	}

	if banner != nil {
		objName, bucket, err := s.uploadImage(ctx, banner, "community_banner")
		if err != nil {
			return nil, nil, fmt.Errorf("banner upload failed: %w", err)
		}
		community.BannerObjectName = objName
		community.BannerBucket = bucket
	}

	var savedRules []model.CommunityRule
	if err := s.communityRepo.Transaction(func(tx *gorm.DB) error {
		if err := s.communityRepo.SaveWithTx(tx, community); err != nil {
			return err
		}

		member := &model.CommunityMember{
			CommunityID: community.CommunityID,
			UserID:      userID,
			Role:        model.CommunityMemberRoleOwner,
		}
		if err := s.memberRepo.SaveWithTx(tx, member); err != nil {
			return err
		}

		savedRules = buildCommunityRules(community.CommunityID, cleanedRules)
		return s.ruleRepo.BulkCreateWithTx(tx, savedRules)
	}); err != nil {
		return nil, nil, err
	}

	invalidateDiscoverCommunitiesCache(ctx, s.cacheRepo)
	s.hydrateCommunityCategory(community)
	return community, savedRules, nil
}

func (s *CommunityService) Update(ctx context.Context, id, userID uint, name, description string, keepRuleIDs []uint, newRules []string, categoryID *uuid.UUID, visibility model.CommunityVisibility, banner *multipart.FileHeader, clearBanner bool) (*model.Community, []model.CommunityRule, error) {
	community, err := s.communityRepo.FindByID(id)
	if err != nil {
		return nil, nil, errors.New("community not found")
	}
	if !community.IsWritable() {
		return nil, nil, errors.New("community is inactive")
	}
	actor, err := s.memberRepo.FindMember(id, userID)
	if err != nil || !isAdminOrAbove(actor.Role) {
		return nil, nil, errors.New("forbidden")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil, errors.New("name is required")
	}
	if err := s.ensureCommunityNameAvailable(name, &id); err != nil {
		return nil, nil, err
	}
	if visibility != model.CommunityVisibilityPublic && visibility != model.CommunityVisibilityPrivate {
		visibility = model.CommunityVisibilityPublic
	}
	if err := s.validateCategoryID(categoryID); err != nil {
		return nil, nil, err
	}
	dedupedKeepIDs := dedupeRuleIDs(keepRuleIDs)
	cleanedNewRules, err := sanitizeCommunityRules(newRules)
	if err != nil {
		return nil, nil, err
	}
	if len(dedupedKeepIDs)+len(cleanedNewRules) > maxCommunityRules {
		return nil, nil, errors.New("maximum 10 rules allowed")
	}
	if len(dedupedKeepIDs) > 0 {
		existing, err := s.ruleRepo.ListByIDsAndCommunity(community.CommunityID, dedupedKeepIDs)
		if err != nil {
			return nil, nil, err
		}
		if len(existing) != len(dedupedKeepIDs) {
			return nil, nil, errors.New("one or more keep_rule_ids do not belong to community")
		}
	}

	community.Name = name
	community.Description = description
	community.CategoryID = categoryID
	community.Visibility = visibility

	if clearBanner && community.BannerObjectName != "" {
		_ = s.storageUsecase.Delete(ctx, community.BannerBucket, community.BannerObjectName)
		community.BannerObjectName = ""
		community.BannerBucket = ""
	} else if banner != nil {
		if community.BannerObjectName != "" {
			_ = s.storageUsecase.Delete(ctx, community.BannerBucket, community.BannerObjectName)
		}
		objName, bucket, err := s.uploadImage(ctx, banner, "community_banner")
		if err != nil {
			return nil, nil, fmt.Errorf("banner upload failed: %w", err)
		}
		community.BannerObjectName = objName
		community.BannerBucket = bucket
	}

	var savedRules []model.CommunityRule
	if err := s.communityRepo.Transaction(func(tx *gorm.DB) error {
		if err := s.communityRepo.UpdateWithTx(tx, community); err != nil {
			return err
		}
		if err := s.ruleRepo.DeleteByCommunityNotInWithTx(tx, community.CommunityID, dedupedKeepIDs); err != nil {
			return err
		}
		// Two-phase reorder: shift kept rules to negative temp order to avoid unique idx conflict.
		for i, ruleID := range dedupedKeepIDs {
			if err := s.ruleRepo.UpdateOrderWithTx(tx, ruleID, -(i + 1)); err != nil {
				return err
			}
		}
		for i, ruleID := range dedupedKeepIDs {
			if err := s.ruleRepo.UpdateOrderWithTx(tx, ruleID, i+1); err != nil {
				return err
			}
		}
		newRuleRecords := make([]model.CommunityRule, len(cleanedNewRules))
		for i, text := range cleanedNewRules {
			newRuleRecords[i] = model.CommunityRule{
				CommunityID: community.CommunityID,
				OrderNum:    len(dedupedKeepIDs) + i + 1,
				Text:        text,
			}
		}
		return s.ruleRepo.BulkCreateWithTx(tx, newRuleRecords)
	}); err != nil {
		return nil, nil, err
	}
	final, err := s.ruleRepo.ListByCommunity(community.CommunityID)
	if err != nil {
		return nil, nil, err
	}
	savedRules = final

	invalidateDiscoverCommunitiesCache(ctx, s.cacheRepo)
	s.hydrateCommunityCategory(community)
	return community, savedRules, nil
}

func (s *CommunityService) Delete(ctx context.Context, id, userID uint) error {
	community, err := s.communityRepo.FindByID(id)
	if err != nil {
		return errors.New("community not found")
	}
	actor, err := s.memberRepo.FindMember(id, userID)
	if err != nil || !isAdminOrAbove(actor.Role) {
		return errors.New("forbidden")
	}
	if community.BannerObjectName != "" {
		_ = s.storageUsecase.Delete(ctx, community.BannerBucket, community.BannerObjectName)
	}
	if err := s.communityRepo.Delete(id); err != nil {
		return err
	}
	invalidateDiscoverCommunitiesCache(ctx, s.cacheRepo)
	return nil
}

func (s *CommunityService) Discover(ctx context.Context, categoryID *uuid.UUID) ([]model.DiscoverCommunity, error) {
	if categoryID != nil && *categoryID != uuid.Nil {
		if _, err := s.categoryRepo.FindByID(*categoryID); err != nil {
			return nil, errors.New("categoryId doesn't exist")
		}
	}

	var cached []model.DiscoverCommunity
	if s.cacheRepo != nil {
		if err := s.cacheRepo.Get(ctx, discoverCommunitiesCacheKey, &cached); err == nil {
			return filterDiscoverByCategory(cached, categoryID), nil
		}
	}

	communities, err := s.communityRepo.FindDiscover()
	if err != nil {
		return nil, err
	}
	if s.cacheRepo != nil {
		_ = s.cacheRepo.Set(ctx, discoverCommunitiesCacheKey, communities, discoverCommunitiesTTL)
	}
	return filterDiscoverByCategory(communities, categoryID), nil
}

func (s *CommunityService) SearchByName(ctx context.Context, name string) ([]model.DiscoverCommunity, error) {
	name = strings.TrimSpace(name)
	var cached []model.DiscoverCommunity
	if s.cacheRepo != nil {
		if err := s.cacheRepo.Get(ctx, communitySearchCacheKey(name), &cached); err == nil {
			return cached, nil
		}
	}

	communities, err := s.communityRepo.FindDiscoverByName(name)
	if err != nil {
		return nil, err
	}
	if s.cacheRepo != nil {
		_ = s.cacheRepo.Set(ctx, communitySearchCacheKey(name), communities, discoverCommunitiesTTL)
	}
	return communities, nil
}

func filterDiscoverByCategory(communities []model.DiscoverCommunity, categoryID *uuid.UUID) []model.DiscoverCommunity {
	if categoryID == nil || *categoryID == uuid.Nil {
		return communities
	}
	out := communities[:0:0]
	for _, c := range communities {
		if c.Category != nil && c.Category.ID == *categoryID {
			out = append(out, c)
		}
	}
	return out
}

func (s *CommunityService) GetByID(id uint) (*model.Community, error) {
	community, err := s.communityRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	s.hydrateCommunityCategory(community)
	return community, nil
}

func (s *CommunityService) GetMyCreated(userID uint, page, limit int) ([]model.Community, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	communities, total, err := s.communityRepo.FindByCreatorID(userID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	s.hydrateCommunityCategories(communities)
	return communities, total, nil
}

func (s *CommunityService) GetMyJoined(userID uint, page, limit int) ([]model.Community, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	communities, total, err := s.communityRepo.FindJoinedByUserID(userID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	s.hydrateCommunityCategories(communities)
	return communities, total, nil
}

func (s *CommunityService) GetURL(bucket, objectName string) string {
	return s.storageUsecase.GetURL(bucket, objectName)
}

func (s *CommunityService) uploadImage(ctx context.Context, header *multipart.FileHeader, uploadType string) (string, string, error) {
	return helper.UploadImage(ctx, s.storageUsecase, header, uploadType)
}

func (s *CommunityService) validateCategoryID(categoryID *uuid.UUID) error {
	if categoryID == nil || *categoryID == uuid.Nil {
		return errors.New("category is required")
	}
	if _, err := s.categoryRepo.FindByID(*categoryID); err != nil {
		return errors.New("category not found")
	}
	return nil
}

func (s *CommunityService) ensureCommunityNameAvailable(name string, excludeID *uint) error {
	exists, err := s.communityRepo.ExistsByName(name, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("community name already taken")
	}
	return nil
}

func (s *CommunityService) hydrateCommunityCategories(communities []model.Community) {
	for i := range communities {
		s.hydrateCommunityCategory(&communities[i])
	}
}

func (s *CommunityService) hydrateCommunityCategory(community *model.Community) {
	if community == nil || community.CategoryID == nil {
		if community != nil {
			community.Category = nil
		}
		return
	}
	category, err := s.categoryRepo.FindByID(*community.CategoryID)
	if err != nil {
		community.Category = nil
		return
	}
	community.Category = category
}
