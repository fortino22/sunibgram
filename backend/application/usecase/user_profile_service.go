package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

func userAvatarCacheKey(userID uint) string {
	return fmt.Sprintf("user:%d:avatar", userID)
}

type UserProfileService struct {
	userRepo        repository.IUserRepository
	followRepo      repository.IFollowRepository
	profileRepo     repository.IUserProfileRepository
	userFacultyRepo repository.IUserFacultyRepository
	cacheRepo       repository.ICacheRepository
	storageUsecase  *StorageUsecase
}

func NewUserProfileService(
	userRepo repository.IUserRepository,
	followRepo repository.IFollowRepository,
	profileRepo repository.IUserProfileRepository,
	userFacultyRepo repository.IUserFacultyRepository,
	cacheRepo repository.ICacheRepository,
	storageUsecase *StorageUsecase,
) service.IUserProfileService {
	return &UserProfileService{
		userRepo:        userRepo,
		followRepo:      followRepo,
		profileRepo:     profileRepo,
		userFacultyRepo: userFacultyRepo,
		cacheRepo:       cacheRepo,
		storageUsecase:  storageUsecase,
	}
}

func (s *UserProfileService) AvatarCacheRepository() repository.ICacheRepository {
	return s.cacheRepo
}

func (s *UserProfileService) UpsertProfile(ctx context.Context, userID uint, input service.UpsertProfileInput) error {
	if userID == 0 {
		return errors.New("invalid user id")
	}
	existing, _ := s.profileRepo.FindByUserID(userID)
	if existing == nil {
		existing = &model.UserProfile{UserID: userID}
	}

	if input.Avatar != nil {
		if err := model.ValidateImageExt(input.Avatar.Filename); err != nil {
			return err
		}
		bucket, err := s.storageUsecase.ResolveBucket("avatar")
		if err != nil {
			return err
		}
		if existing.AvatarObjectName != "" && existing.AvatarBucket != "" {
			if err := s.storageUsecase.Delete(ctx, existing.AvatarBucket, existing.AvatarObjectName); err != nil {
				return fmt.Errorf("failed to delete old avatar: %w", err)
			}
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
		if err := model.ValidateImageExt(input.Banner.Filename); err != nil {
			return err
		}
		bucket, err := s.storageUsecase.ResolveBucket("banner")
		if err != nil {
			return err
		}
		if existing.BannerObjectName != "" && existing.BannerBucket != "" {
			if err := s.storageUsecase.Delete(ctx, existing.BannerBucket, existing.BannerObjectName); err != nil {
				return fmt.Errorf("failed to delete old banner: %w", err)
			}
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

	if err := s.profileRepo.Upsert(existing); err != nil {
		return err
	}
	if input.Avatar != nil {
		invalidateDiscoverCommunitiesCache(ctx, s.cacheRepo)
	}
	s.invalidateUserProfileCaches(ctx, userID)
	return nil
}

func (s *UserProfileService) UpdateProfile(ctx context.Context, userID uint, input service.UpdateProfileInput) error {
	if input.Name != nil || input.Username != nil {
		user, err := s.userRepo.FindByID(userID)
		if err != nil {
			return errors.New("user not found")
		}
		name := user.Name
		username := user.UserName
		if input.Name != nil {
			name = strings.TrimSpace(*input.Name)
			if name == "" {
				return errors.New("name cannot be empty")
			}
		}
		if input.Username != nil {
			username = strings.TrimSpace(*input.Username)
			if username == "" {
				return errors.New("username cannot be empty")
			}
			if username != user.UserName {
				taken, err := s.userRepo.ExistsByUserName(username)
				if err != nil {
					return err
				}
				if taken {
					return errors.New("username already taken")
				}
			}
		}
		if err := s.userRepo.UpdateInfo(userID, name, username); err != nil {
			return err
		}
		if s.cacheRepo != nil {
			_ = s.cacheRepo.DeleteByPattern(ctx, "community:threads:top:*")
			_ = s.cacheRepo.DeleteByPattern(ctx, "community:threads:latest:*")
			_ = s.cacheRepo.DeleteByPattern(ctx, "thread:all-communities:*")

			// Invalidate all feed caches that embed the user's name inside thread objects.
			_ = s.cacheRepo.DeleteByPattern(ctx, fmt.Sprintf("thread:userfeed:%d:*", userID))
			_ = s.cacheRepo.DeleteByPattern(ctx, fmt.Sprintf("thread:usermedia:%d:*", userID))
			_ = s.cacheRepo.DeleteByPattern(ctx, fmt.Sprintf("feed:foryou:%d:*", userID))
			_ = s.cacheRepo.Delete(ctx, "feed:trending:global")
			_ = s.cacheRepo.DeleteByPattern(ctx, "thread:following:*")
		}
		s.invalidateUserProfileCaches(ctx, userID)
	}

	if input.ClearAvatar || input.ClearBanner {
		existing, _ := s.profileRepo.FindByUserID(userID)
		if existing != nil {
			if input.ClearAvatar && existing.AvatarObjectName != "" && existing.AvatarBucket != "" {
				if err := s.storageUsecase.Delete(ctx, existing.AvatarBucket, existing.AvatarObjectName); err != nil {
					return fmt.Errorf("failed to delete avatar: %w", err)
				}
				existing.AvatarObjectName = ""
				existing.AvatarBucket = ""
			}
			if input.ClearBanner && existing.BannerObjectName != "" && existing.BannerBucket != "" {
				if err := s.storageUsecase.Delete(ctx, existing.BannerBucket, existing.BannerObjectName); err != nil {
					return fmt.Errorf("failed to delete banner: %w", err)
				}
				existing.BannerObjectName = ""
				existing.BannerBucket = ""
			}
			if err := s.profileRepo.Upsert(existing); err != nil {
				return err
			}
			if input.ClearAvatar {
				invalidateDiscoverCommunitiesCache(ctx, s.cacheRepo)
			}
			s.invalidateUserProfileCaches(ctx, userID)
		}
	}

	if input.BirthdayDate != nil {
		if err := model.ValidateBirthdayDate(*input.BirthdayDate); err != nil {
			return err
		}
		existing, _ := s.profileRepo.FindByUserID(userID)
		if existing == nil {
			return errors.New("profile not found")
		}
		existing.BirthdayDate = *input.BirthdayDate
		if err := s.profileRepo.Upsert(existing); err != nil {
			return err
		}
		s.invalidateUserProfileCaches(ctx, userID)
	}

	mediaInput := service.UpsertProfileInput{
		Avatar: input.Avatar,
		Banner: input.Banner,
		Bio:    input.Bio,
	}
	if mediaInput.Avatar != nil || mediaInput.Banner != nil || mediaInput.Bio != nil {
		return s.UpsertProfile(ctx, userID, mediaInput)
	}
	return nil
}

func (s *UserProfileService) invalidateUserProfileCaches(ctx context.Context, userID uint) {
	if s.cacheRepo == nil {
		return
	}
	_ = s.cacheRepo.Delete(ctx,
		userAvatarCacheKey(userID),
		fmt.Sprintf("user:id:%d", userID),
		fmt.Sprintf("profile:user:%d", userID),
		fmt.Sprintf("notification:sender:%d", userID),
		fmt.Sprintf("userfaculty:user:%d", userID),
	)
	_ = s.cacheRepo.Delete(ctx, "reports:all")
	_ = s.cacheRepo.DeleteByPattern(ctx, "reports:id:*")
	_ = s.cacheRepo.DeleteByPattern(ctx, fmt.Sprintf("thread:userfeed:%d:*", userID))
	_ = s.cacheRepo.DeleteByPattern(ctx, fmt.Sprintf("thread:usermedia:%d:*", userID))
	_ = s.cacheRepo.DeleteByPattern(ctx, fmt.Sprintf("feed:foryou:%d:*", userID))
	_ = s.cacheRepo.DeleteByPattern(ctx, fmt.Sprintf("bookmarks:user:%d:*", userID))
	_ = s.cacheRepo.DeleteByPattern(ctx, fmt.Sprintf("likes:user:%d:*", userID))
	_ = s.cacheRepo.DeleteByPattern(ctx, fmt.Sprintf("thread:my-communities:user:%d:*", userID))
	_ = s.cacheRepo.DeleteByPattern(ctx, "thread:roots:*")
	_ = s.cacheRepo.DeleteByPattern(ctx, "thread:following:*")
	_ = s.cacheRepo.DeleteByPattern(ctx, "thread:community:*")
	_ = s.cacheRepo.DeleteByPattern(ctx, "thread:all-communities:*")
	_ = s.cacheRepo.DeleteByPattern(ctx, "event:threads:*")
	_ = s.cacheRepo.Delete(ctx, "feed:trending:global")
	_ = s.cacheRepo.DeleteByPattern(ctx, "community:requests:*")
}

func (s *UserProfileService) GetProfile(userID uint, viewerID *uint) (*service.FullProfile, error) {
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
	userFaculty, _ := s.userFacultyRepo.FindByUserID(userID)
	isFollowing := false
	if viewerID != nil && *viewerID != 0 && *viewerID != userID {
		isFollowing, err = s.followRepo.Exists(*viewerID, userID)
		if err != nil {
			return nil, err
		}
	}
	return &service.FullProfile{
		User:        user,
		Stat:        stat,
		Profile:     profile,
		UserFaculty: userFaculty,
		IsFollowing: isFollowing,
	}, nil
}

func (s *UserProfileService) GetProfiles(userIDs []uint) ([]model.UserProfile, error) {
	return s.profileRepo.FindByUserIDs(userIDs)
}

func (s *UserProfileService) GetUsers(userIDs []uint) ([]model.User, error) {
	return s.userRepo.FindByIDs(userIDs)
}

func (s *UserProfileService) GetUserFaculties(userIDs []uint) ([]model.UserFaculty, error) {
	return s.userFacultyRepo.FindByUserIDs(userIDs)
}

func (s *UserProfileService) GetURL(bucket, objectName string) string {
	return s.storageUsecase.GetURL(bucket, objectName)
}
