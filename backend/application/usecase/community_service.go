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
	cacheRepo      repository.ICacheRepository
	storageUsecase *StorageUsecase
}

func NewCommunityService(
	communityRepo repository.ICommunityRepository,
	memberRepo repository.ICommunityMemberRepository,
	userRepo repository.IUserRepository,
	categoryRepo repository.ICategoryRepository,
	cacheRepo repository.ICacheRepository,
	storageUsecase *StorageUsecase,
) service.ICommunityService {
	return &CommunityService{
		communityRepo:  communityRepo,
		memberRepo:     memberRepo,
		userRepo:       userRepo,
		categoryRepo:   categoryRepo,
		cacheRepo:      cacheRepo,
		storageUsecase: storageUsecase,
	}
}

func (s *CommunityService) Create(ctx context.Context, userID uint, name, description, rules string, categoryID *uuid.UUID, visibility model.CommunityVisibility, avatar, banner *multipart.FileHeader) (*model.Community, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if !user.IsPremium {
		return nil, errors.New("only premium users can create communities")
	}

	if strings.TrimSpace(name) == "" {
		return nil, errors.New("name is required")
	}
	if visibility != model.CommunityVisibilityPublic && visibility != model.CommunityVisibilityPrivate {
		visibility = model.CommunityVisibilityPublic
	}
	if err := s.validateCategoryID(categoryID); err != nil {
		return nil, err
	}

	community := &model.Community{
		UserID:      userID,
		Name:        name,
		Description: description,
		Rules:       rules,
		CategoryID:  categoryID,
		Visibility:  visibility,
		MemberCount: 1,
	}

	if avatar != nil {
		objName, bucket, err := s.uploadImage(ctx, avatar, "community_avatar")
		if err != nil {
			return nil, fmt.Errorf("avatar upload failed: %w", err)
		}
		community.AvatarObjectName = objName
		community.AvatarBucket = bucket
	}

	if banner != nil {
		objName, bucket, err := s.uploadImage(ctx, banner, "community_banner")
		if err != nil {
			return nil, fmt.Errorf("banner upload failed: %w", err)
		}
		community.BannerObjectName = objName
		community.BannerBucket = bucket
	}

	if err := s.communityRepo.Transaction(func(tx *gorm.DB) error {
		if err := s.communityRepo.SaveWithTx(tx, community); err != nil {
			return err
		}

		member := &model.CommunityMember{
			CommunityID: community.CommunityID,
			UserID:      userID,
			Role:        model.CommunityMemberRoleOwner,
		}
		return s.memberRepo.SaveWithTx(tx, member)
	}); err != nil {
		return nil, err
	}

	invalidateDiscoverCommunitiesCache(ctx, s.cacheRepo)
	s.hydrateCommunityCategory(community)
	return community, nil
}

func (s *CommunityService) Update(ctx context.Context, id, userID uint, name, description, rules string, categoryID *uuid.UUID, visibility model.CommunityVisibility, avatar, banner *multipart.FileHeader, clearAvatar, clearBanner bool) (*model.Community, error) {
	community, err := s.communityRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("community not found")
	}
	if !community.IsWritable() {
		return nil, errors.New("community is inactive")
	}
	actor, err := s.memberRepo.FindMember(id, userID)
	if err != nil || !isAdminOrAbove(actor.Role) {
		return nil, errors.New("forbidden")
	}
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("name is required")
	}
	if visibility != model.CommunityVisibilityPublic && visibility != model.CommunityVisibilityPrivate {
		visibility = model.CommunityVisibilityPublic
	}
	if err := s.validateCategoryID(categoryID); err != nil {
		return nil, err
	}

	community.Name = name
	community.Description = description
	community.Rules = rules
	community.CategoryID = categoryID
	community.Visibility = visibility

	if clearAvatar && community.AvatarObjectName != "" {
		_ = s.storageUsecase.Delete(ctx, community.AvatarBucket, community.AvatarObjectName)
		community.AvatarObjectName = ""
		community.AvatarBucket = ""
	} else if avatar != nil {
		if community.AvatarObjectName != "" {
			_ = s.storageUsecase.Delete(ctx, community.AvatarBucket, community.AvatarObjectName)
		}
		objName, bucket, err := s.uploadImage(ctx, avatar, "community_avatar")
		if err != nil {
			return nil, fmt.Errorf("avatar upload failed: %w", err)
		}
		community.AvatarObjectName = objName
		community.AvatarBucket = bucket
	}

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
			return nil, fmt.Errorf("banner upload failed: %w", err)
		}
		community.BannerObjectName = objName
		community.BannerBucket = bucket
	}

	if err := s.communityRepo.Update(community); err != nil {
		return nil, err
	}
	invalidateDiscoverCommunitiesCache(ctx, s.cacheRepo)
	s.hydrateCommunityCategory(community)
	return community, nil
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
	if community.AvatarObjectName != "" {
		_ = s.storageUsecase.Delete(ctx, community.AvatarBucket, community.AvatarObjectName)
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
