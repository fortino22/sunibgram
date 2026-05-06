package usecase

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type CommunityService struct {
	communityRepo  repository.ICommunityRepository
	memberRepo     repository.ICommunityMemberRepository
	userRepo       repository.IUserRepository
	storageUsecase *StorageUsecase
}

func NewCommunityService(
	communityRepo repository.ICommunityRepository,
	memberRepo repository.ICommunityMemberRepository,
	userRepo repository.IUserRepository,
	storageUsecase *StorageUsecase,
) service.ICommunityService {
	return &CommunityService{
		communityRepo:  communityRepo,
		memberRepo:     memberRepo,
		userRepo:       userRepo,
		storageUsecase: storageUsecase,
	}
}

func (s *CommunityService) Create(ctx context.Context, userID uint, name, description, rules, category string, visibility model.CommunityVisibility, avatar, banner *multipart.FileHeader) (*model.Community, error) {
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
	if strings.TrimSpace(category) == "" {
		return nil, errors.New("category is required")
	}
	if visibility != model.CommunityVisibilityPublic && visibility != model.CommunityVisibilityPrivate {
		visibility = model.CommunityVisibilityPublic
	}

	community := &model.Community{
		UserID:      userID,
		Name:        name,
		Description: description,
		Rules:       rules,
		Category:    strings.ToLower(strings.TrimSpace(category)),
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

	if err := s.communityRepo.Save(community); err != nil {
		return nil, err
	}

	member := &model.CommunityMember{
		CommunityID: community.CommunityID,
		UserID:      userID,
		Role:        model.CommunityMemberRoleOwner,
	}
	if err := s.memberRepo.Save(member); err != nil {
		return nil, err
	}

	return community, nil
}

func (s *CommunityService) Update(ctx context.Context, id, userID uint, name, description, rules, category string, visibility model.CommunityVisibility, avatar, banner *multipart.FileHeader, clearAvatar, clearBanner bool) (*model.Community, error) {
	community, err := s.communityRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("community not found")
	}
	actor, err := s.memberRepo.FindMember(id, userID)
	if err != nil || !isAdminOrAbove(actor.Role) {
		return nil, errors.New("forbidden")
	}
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("name is required")
	}
	if strings.TrimSpace(category) == "" {
		return nil, errors.New("category is required")
	}
	if visibility != model.CommunityVisibilityPublic && visibility != model.CommunityVisibilityPrivate {
		visibility = model.CommunityVisibilityPublic
	}

	community.Name = name
	community.Description = description
	community.Rules = rules
	community.Category = strings.ToLower(strings.TrimSpace(category))
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
	return s.communityRepo.Delete(id)
}

func (s *CommunityService) GetByID(id uint) (*model.Community, error) {
	return s.communityRepo.FindByID(id)
}

func (s *CommunityService) GetAll(page, limit int, category string) ([]model.Community, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.communityRepo.FindAll(page, limit, category)
}

func (s *CommunityService) GetMyCreated(userID uint, page, limit int) ([]model.Community, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.communityRepo.FindByCreatorID(userID, page, limit)
}

func (s *CommunityService) GetMyJoined(userID uint, page, limit int) ([]model.Community, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.communityRepo.FindJoinedByUserID(userID, page, limit)
}

func (s *CommunityService) GetURL(bucket, objectName string) string {
	return s.storageUsecase.GetURL(bucket, objectName)
}

func (s *CommunityService) uploadImage(ctx context.Context, header *multipart.FileHeader, uploadType string) (string, string, error) {
	if err := model.ValidateImageExt(header.Filename); err != nil {
		return "", "", err
	}
	bucket, err := s.storageUsecase.ResolveBucket(uploadType)
	if err != nil {
		return "", "", err
	}
	f, err := header.Open()
	if err != nil {
		return "", "", err
	}
	defer f.Close()
	objName, err := s.storageUsecase.Upload(ctx, bucket, f, header)
	if err != nil {
		return "", "", err
	}
	return objName, bucket, nil
}
