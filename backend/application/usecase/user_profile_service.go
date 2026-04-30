package usecase

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

var allowedImageExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
}

type UserProfileService struct {
	userRepo       repository.IUserRepository
	followRepo     repository.IFollowRepository
	profileRepo    repository.IUserProfileRepository
	storageUsecase *StorageUsecase
}

func NewUserProfileService(
	userRepo repository.IUserRepository,
	followRepo repository.IFollowRepository,
	profileRepo repository.IUserProfileRepository,
	storageUsecase *StorageUsecase,
) service.IUserProfileService {
	return &UserProfileService{
		userRepo:       userRepo,
		followRepo:     followRepo,
		profileRepo:    profileRepo,
		storageUsecase: storageUsecase,
	}
}

func (s *UserProfileService) UpsertProfile(ctx context.Context, userID uint, input service.UpsertProfileInput) error {
	existing, _ := s.profileRepo.FindByUserID(userID)
	if existing == nil {
		existing = &model.UserProfile{UserID: userID}
	}

	if input.Avatar != nil {
		ext := strings.ToLower(filepath.Ext(input.Avatar.Filename))
		if !allowedImageExts[ext] {
			return errors.New("avatar must be jpg, jpeg, png, or webp")
		}
		bucket, err := s.storageUsecase.ResolveBucket("avatar")
		if err != nil {
			return err
		}
		if existing.AvatarObjectName != "" {
			_ = s.storageUsecase.Delete(ctx, existing.AvatarBucket, existing.AvatarObjectName)
		}
		f, err := input.Avatar.Open()
		if err != nil {
			return err
		}
		objectName, err := s.storageUsecase.Upload(ctx, bucket, f, input.Avatar)
		f.Close()
		if err != nil {
			return err
		}
		existing.AvatarObjectName = objectName
		existing.AvatarBucket = bucket
	}

	if input.Banner != nil {
		ext := strings.ToLower(filepath.Ext(input.Banner.Filename))
		if !allowedImageExts[ext] {
			return errors.New("banner must be jpg, jpeg, png, or webp")
		}
		bucket, err := s.storageUsecase.ResolveBucket("banner")
		if err != nil {
			return err
		}
		if existing.BannerObjectName != "" {
			_ = s.storageUsecase.Delete(ctx, existing.BannerBucket, existing.BannerObjectName)
		}
		f, err := input.Banner.Open()
		if err != nil {
			return err
		}
		objectName, err := s.storageUsecase.Upload(ctx, bucket, f, input.Banner)
		f.Close()
		if err != nil {
			return err
		}
		existing.BannerObjectName = objectName
		existing.BannerBucket = bucket
	}

	if input.Bio != nil {
		existing.Bio = *input.Bio
	}

	return s.profileRepo.Upsert(existing)
}

func (s *UserProfileService) GetProfile(userID uint) (*service.FullProfile, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	stat, err := s.followRepo.GetStat(userID)
	if err != nil {
		return nil, err
	}
	profile, _ := s.profileRepo.FindByUserID(userID)
	if profile == nil {
		profile = &model.UserProfile{UserID: userID}
	}
	return &service.FullProfile{User: user, Stat: stat, Profile: profile}, nil
}

func (s *UserProfileService) GetURL(bucket, objectName string) string {
	return s.storageUsecase.GetURL(bucket, objectName)
}
