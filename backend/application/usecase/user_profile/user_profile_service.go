package userprofileuc

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/sunibgram/application/helper"
	communityuc "github.com/sunibgram/application/usecase/community"
	storageuc "github.com/sunibgram/application/usecase/storage"
	mediamodel "github.com/sunibgram/domain/model/media"
	profilefieldchangemodel "github.com/sunibgram/domain/model/profile_field_change"
	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
	cacherepoiface "github.com/sunibgram/domain/repository/cache"
	facultylocationrepoiface "github.com/sunibgram/domain/repository/faculty_location"
	followrepoiface "github.com/sunibgram/domain/repository/follow"
	notificationusermuterepoiface "github.com/sunibgram/domain/repository/notification_user_mute"
	profilefieldchangerepoiface "github.com/sunibgram/domain/repository/profile_field_change"
	userrepoiface "github.com/sunibgram/domain/repository/user"
	userblockrepoiface "github.com/sunibgram/domain/repository/user_block"
	userfacultyrepoiface "github.com/sunibgram/domain/repository/user_faculty"
	userprofilerepoiface "github.com/sunibgram/domain/repository/user_profile"
	userprofilesvciface "github.com/sunibgram/domain/service/user_profile"
)

func userAvatarCacheKey(userID uint) string {
	return fmt.Sprintf("user:%d:avatar", userID)
}

type UserProfileService struct {
	userRepo            userrepoiface.IUserRepository
	followRepo          followrepoiface.IFollowRepository
	blockRepo           userblockrepoiface.IUserBlockRepository
	muteRepo            notificationusermuterepoiface.INotificationUserMuteRepository
	profileRepo         userprofilerepoiface.IUserProfileRepository
	pfcRepo             profilefieldchangerepoiface.IProfileFieldChangeRepository
	userFacultyRepo     userfacultyrepoiface.IUserFacultyRepository
	facultyLocationRepo facultylocationrepoiface.IFacultyLocationRepository
	cacheRepo           cacherepoiface.ICacheRepository
	storageUsecase      *storageuc.StorageUsecase
}

func NewUserProfileService(
	userRepo userrepoiface.IUserRepository,
	followRepo followrepoiface.IFollowRepository,
	blockRepo userblockrepoiface.IUserBlockRepository,
	muteRepo notificationusermuterepoiface.INotificationUserMuteRepository,
	profileRepo userprofilerepoiface.IUserProfileRepository,
	pfcRepo profilefieldchangerepoiface.IProfileFieldChangeRepository,
	userFacultyRepo userfacultyrepoiface.IUserFacultyRepository,
	facultyLocationRepo facultylocationrepoiface.IFacultyLocationRepository,
	cacheRepo cacherepoiface.ICacheRepository,
	storageUsecase *storageuc.StorageUsecase,
) userprofilesvciface.IUserProfileService {
	return &UserProfileService{
		userRepo:            userRepo,
		followRepo:          followRepo,
		blockRepo:           blockRepo,
		muteRepo:            muteRepo,
		profileRepo:         profileRepo,
		pfcRepo:             pfcRepo,
		userFacultyRepo:     userFacultyRepo,
		facultyLocationRepo: facultyLocationRepo,
		cacheRepo:           cacheRepo,
		storageUsecase:      storageUsecase,
	}
}

func (s *UserProfileService) AvatarCacheRepository() cacherepoiface.ICacheRepository {
	return s.cacheRepo
}

func (s *UserProfileService) UpsertProfile(ctx context.Context, userID uint, input userprofilesvciface.UpsertProfileInput) error {
	if userID == 0 {
		return errors.New("invalid user id")
	}
	existing, _ := s.profileRepo.FindByUserID(userID)
	if existing == nil {
		existing = &userprofilemodel.UserProfile{UserID: userID}
	}

	if input.Avatar != nil {
		if err := validateProfileImage(input.Avatar, "avatar"); err != nil {
			return err
		}
		if existing.AvatarObjectName != "" {
			if err := s.storageUsecase.Delete(ctx, existing.AvatarObjectName); err != nil {
				return fmt.Errorf("failed to delete old avatar: %w", err)
			}
		}
		objectName, err := helper.UploadImage(ctx, s.storageUsecase, input.Avatar, "avatar")
		if err != nil {
			return err
		}
		existing.AvatarObjectName = objectName
	}

	if input.Banner != nil {
		if err := validateProfileImage(input.Banner, "banner"); err != nil {
			return err
		}
		if existing.BannerObjectName != "" {
			if err := s.storageUsecase.Delete(ctx, existing.BannerObjectName); err != nil {
				return fmt.Errorf("failed to delete old banner: %w", err)
			}
		}
		objectName, err := helper.UploadImage(ctx, s.storageUsecase, input.Banner, "banner")
		if err != nil {
			return err
		}
		existing.BannerObjectName = objectName
	}

	if input.Bio != nil {
		trimmed := strings.TrimSpace(*input.Bio)
		if utf8.RuneCountInString(trimmed) > 500 {
			return errors.New("bio must be 500 characters or less")
		}
		*input.Bio = trimmed
		existing.Bio = trimmed
	}

	if err := s.profileRepo.Upsert(existing); err != nil {
		return err
	}
	if input.Avatar != nil {
		communityuc.InvalidateDiscoverCommunitiesCache(ctx, s.cacheRepo)
	}
	s.invalidateUserProfileCaches(ctx, userID)
	return nil
}

func (s *UserProfileService) UpdateProfile(ctx context.Context, userID uint, input userprofilesvciface.UpdateProfileInput) error {
	if input.Name != nil || input.Username != nil {
		user, err := s.userRepo.FindByID(userID)
		if err != nil {
			return errors.New("user not found")
		}
		name := user.Name
		username := user.UserName
		if input.Name != nil {
			name = strings.TrimSpace(*input.Name)
			if err := usermodel.ValidateName(name); err != nil {
				return err
			}
		}
		if input.Username != nil {
			username = strings.TrimSpace(*input.Username)
			if err := usermodel.ValidateUsername(username); err != nil {
				return err
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
		// Per-field change cooldown, backed by the profile_field_changes log.
		// Only an actual value change is rate-limited and logged, so re-submitting
		// an unchanged name/username (e.g. editing only bio/avatar) is a no-op.
		now := time.Now()
		latest, _ := s.pfcRepo.LatestByUser(userID) // nil-safe; absent field reads as zero time
		nameChanged := input.Name != nil && name != user.Name
		usernameChanged := input.Username != nil && username != user.UserName
		if nameChanged {
			if next := profilefieldchangemodel.NextChangeAt(profilefieldchangemodel.FieldName, latest[profilefieldchangemodel.FieldName]); next.After(now) {
				return fmt.Errorf("display name can only be changed once every 7 days; try again on %s", next.Format("Jan 2, 2006"))
			}
		}
		if usernameChanged {
			if next := profilefieldchangemodel.NextChangeAt(profilefieldchangemodel.FieldUsername, latest[profilefieldchangemodel.FieldUsername]); next.After(now) {
				return fmt.Errorf("username can only be changed once every 7 days; try again on %s", next.Format("Jan 2, 2006"))
			}
		}
		if err := s.userRepo.UpdateInfo(userID, name, username); err != nil {
			return err
		}
		// Log the change(s) after the write succeeds. Best-effort: a failed log
		// only means the next change isn't blocked, never blocks a valid one.
		if nameChanged {
			_ = s.pfcRepo.Record(userID, profilefieldchangemodel.FieldName, now, user.Name, name)
		}
		if usernameChanged {
			_ = s.pfcRepo.Record(userID, profilefieldchangemodel.FieldUsername, now, user.UserName, username)
		}
		if s.cacheRepo != nil {
			// Tag-based invalidation. Global thread feed families cover the
			// legacy patterns "community:threads:top/latest:*",
			// "thread:all-communities:*", "thread:following:*"; userFeeds /
			// userForYou cover the per-user user feeds and foryou source.
			_ = s.cacheRepo.InvalidateUserFeeds(ctx, userID)
			_ = s.cacheRepo.InvalidateUserForYouFeeds(ctx, userID)
			_ = s.cacheRepo.InvalidateGlobalThreadFeeds(ctx)
			_ = s.cacheRepo.Delete(ctx, "feed:trending:global")
		}
		s.invalidateUserProfileCaches(ctx, userID)
	}

	if input.FacultyLocationID != nil {
		existing, err := s.profileRepo.FindByUserID(userID)
		if err != nil || existing == nil {
			return errors.New("profile not found")
		}
		newLoc := *input.FacultyLocationID
		var curLoc uint
		if existing.FacultyLocationID != nil {
			curLoc = *existing.FacultyLocationID
		}
		if newLoc != curLoc {
			// Must be a real campus the user may pick (not the "All Regions" sentinel).
			ok, err := s.facultyLocationRepo.ExistsSelectable(newLoc)
			if err != nil {
				return err
			}
			if !ok {
				return errors.New("faculty location not found")
			}
			now := time.Now()
			latest, _ := s.pfcRepo.LatestByUser(userID)
			if next := profilefieldchangemodel.NextChangeAt(profilefieldchangemodel.FieldFacultyLocation, latest[profilefieldchangemodel.FieldFacultyLocation]); next.After(now) {
				return fmt.Errorf("campus can only be changed once every 6 months; try again on %s", next.Format("Jan 2, 2006"))
			}
			if err := s.profileRepo.UpdateFacultyLocation(userID, &newLoc); err != nil {
				return err
			}
			_ = s.pfcRepo.Record(userID, profilefieldchangemodel.FieldFacultyLocation, now,
				strconv.FormatUint(uint64(curLoc), 10), strconv.FormatUint(uint64(newLoc), 10))
			s.invalidateUserProfileCaches(ctx, userID)
		}
	}

	if input.ClearAvatar || input.ClearBanner {
		existing, _ := s.profileRepo.FindByUserID(userID)
		if existing != nil {
			if input.ClearAvatar && existing.AvatarObjectName != "" {
				if err := s.storageUsecase.Delete(ctx, existing.AvatarObjectName); err != nil {
					return fmt.Errorf("failed to delete avatar: %w", err)
				}
				existing.AvatarObjectName = ""
			}
			if input.ClearBanner && existing.BannerObjectName != "" {
				if err := s.storageUsecase.Delete(ctx, existing.BannerObjectName); err != nil {
					return fmt.Errorf("failed to delete banner: %w", err)
				}
				existing.BannerObjectName = ""
			}
			if err := s.profileRepo.Upsert(existing); err != nil {
				return err
			}
			if input.ClearAvatar {
				communityuc.InvalidateDiscoverCommunitiesCache(ctx, s.cacheRepo)
			}
			s.invalidateUserProfileCaches(ctx, userID)
		}
	}

	if input.BirthdayDate != nil {
		if err := userprofilemodel.ValidateBirthdayDate(*input.BirthdayDate); err != nil {
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

	mediaInput := userprofilesvciface.UpsertProfileInput{
		Avatar: input.Avatar,
		Banner: input.Banner,
		Bio:    input.Bio,
	}
	if mediaInput.Avatar != nil || mediaInput.Banner != nil || mediaInput.Bio != nil {
		return s.UpsertProfile(ctx, userID, mediaInput)
	}
	return nil
}

func validateProfileImage(header *multipart.FileHeader, uploadType string) error {
	if err := mediamodel.ValidateUploadSize(header, mediamodel.MaxImageUploadSize(uploadType)); err != nil {
		return err
	}
	if err := mediamodel.ValidateImageExt(header.Filename); err != nil {
		return err
	}
	f, err := header.Open()
	if err != nil {
		return err
	}
	defer f.Close()
	return mediamodel.ValidateImageFile(f)
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
	_ = s.cacheRepo.Delete(ctx, "reports:all", "feed:trending:global")
	// Tag-based invalidation — covers thread:userfeed/usermedia/following:<uid>
	// (userFeedsTag), feed:foryou:user:<uid>:src/newcount (userForYouTag),
	// thread:my-communities:user:<uid> + per-viewer feed slices
	// (viewerThreadFeedsTag), and the global thread feed families
	// (thread:roots/community/all-communities/following/event:threads, all
	// registered under globalThreadFeedsTag at every Set site).
	_ = s.cacheRepo.InvalidateUserFeeds(ctx, userID)
	_ = s.cacheRepo.InvalidateUserForYouFeeds(ctx, userID)
	_ = s.cacheRepo.InvalidateViewerThreadFeeds(ctx, userID)
	_ = s.cacheRepo.InvalidateGlobalThreadFeeds(ctx)
	// Per-user interaction lists (bookmarks/likes:user:<uid>:*) → tag invalidation.
	_ = s.cacheRepo.InvalidateInteractionUserList(ctx, "bookmarks", userID)
	_ = s.cacheRepo.InvalidateInteractionUserList(ctx, "likes", userID)
	// community:users:* (member listings showing is_following etc) →
	// communityMembersGlobalTag drops every community's user list.
	_ = s.cacheRepo.InvalidateCommunityMembersGlobal(ctx)
	_ = s.cacheRepo.InvalidateReportsID(ctx)
	_ = s.cacheRepo.InvalidateCommunityRequestsGlobal(ctx)
}

func (s *UserProfileService) GetProfile(userID uint, viewerID *uint) (*userprofilesvciface.FullProfile, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	stat, err := s.followRepo.GetStat(userID)
	if err != nil {
		return nil, err
	}
	profile, _ := s.profileRepo.FindByUserID(userID)
	if profile == nil {
		profile = &userprofilemodel.UserProfile{UserID: userID}
	}
	userFaculty, _ := s.userFacultyRepo.FindByUserID(userID)
	isFollowing := false
	isBlocked := false
	isBlockedBy := false
	isMuted := false
	if viewerID != nil && *viewerID != 0 && *viewerID != userID {
		isFollowing, err = s.followRepo.Exists(*viewerID, userID)
		if err != nil {
			return nil, err
		}
		if s.blockRepo != nil {
			// Directional: viewer blocked target → show Unblock UI.
			isBlocked, err = s.blockRepo.BlocksUser(*viewerID, userID)
			if err != nil {
				return nil, err
			}
			// Directional: target blocked viewer → restricted view, no unblock.
			isBlockedBy, err = s.blockRepo.BlocksUser(userID, *viewerID)
			if err != nil {
				return nil, err
			}
		}
		if s.muteRepo != nil {
			// Directional one-way: viewer muted target → show Unmute UI.
			isMuted, err = s.muteRepo.IsMuted(*viewerID, userID)
			if err != nil {
				return nil, err
			}
		}
	}
	fp := &userprofilesvciface.FullProfile{
		User:        user,
		Stat:        stat,
		Profile:     profile,
		UserFaculty: userFaculty,
		IsFollowing: isFollowing,
		IsBlocked:   isBlocked,
		IsBlockedBy: isBlockedBy,
		IsMuted:     isMuted,
	}
	// Per-field change cooldown for the edit UI. Set only while still locked;
	// the controller renders these self-view only.
	if latest, lerr := s.pfcRepo.LatestByUser(userID); lerr == nil {
		now := time.Now()
		if n := profilefieldchangemodel.NextChangeAt(profilefieldchangemodel.FieldName, latest[profilefieldchangemodel.FieldName]); n.After(now) {
			fp.NameChangeAvailableAt = &n
		}
		if n := profilefieldchangemodel.NextChangeAt(profilefieldchangemodel.FieldUsername, latest[profilefieldchangemodel.FieldUsername]); n.After(now) {
			fp.UsernameChangeAvailableAt = &n
		}
		if n := profilefieldchangemodel.NextChangeAt(profilefieldchangemodel.FieldFacultyLocation, latest[profilefieldchangemodel.FieldFacultyLocation]); n.After(now) {
			fp.FacultyLocationChangeAvailableAt = &n
		}
	}
	return fp, nil
}

// onboardingTargetFollows is the upper bound on how many people a first-time
// user is asked to follow. The effective requirement is capped at the number of
// followable (non-admin) users on the app, so small/empty deployments stay
// clearable: required = min(onboardingTargetFollows, followablePool).
const onboardingTargetFollows = 5

// requiredFollows is the dynamic follow requirement for the onboarding gate:
// the target (5) capped at the addressable pool. Pool 0 → 0 (gate is skippable).
func (s *UserProfileService) requiredFollows(userID uint) (int, error) {
	pool, err := s.userRepo.CountFollowablePeers(userID)
	if err != nil {
		return 0, err
	}
	required := onboardingTargetFollows
	if int64(required) > pool {
		required = int(pool)
	}
	return required, nil
}

// OnboardingStatus returns the dynamic follow requirement and the user's current
// following count, used by the frontend to size the onboarding gate.
func (s *UserProfileService) OnboardingStatus(userID uint) (int, int, error) {
	required, err := s.requiredFollows(userID)
	if err != nil {
		return 0, 0, err
	}
	following := 0
	if stat, err := s.followRepo.GetStat(userID); err == nil && stat != nil {
		following = stat.FollowingCount
	}
	return required, following, nil
}

// CompleteOnboarding clears the first-time follow gate. Idempotent (no-op if
// already done) and server-authoritative: rejects unless the user follows at
// least the dynamic requirement, so the UI check cannot be bypassed. When the
// app has fewer than the target followable users (including zero), the
// requirement shrinks accordingly so the gate stays clearable.
func (s *UserProfileService) CompleteOnboarding(userID uint) error {
	if done, err := s.userRepo.IsOnboarded(userID); err == nil && done {
		return nil
	}
	required, err := s.requiredFollows(userID)
	if err != nil {
		return err
	}
	following := 0
	if stat, err := s.followRepo.GetStat(userID); err == nil && stat != nil {
		following = stat.FollowingCount
	}
	if following < required {
		return fmt.Errorf("must follow at least %d people to continue", required)
	}
	return s.userRepo.CompleteOnboarding(userID)
}

// UsernameHistory returns the user's username-change log, newest first, for the
// "about this account" page.
func (s *UserProfileService) UsernameHistory(userID uint) ([]profilefieldchangemodel.ProfileFieldChange, error) {
	return s.pfcRepo.HistoryByUserField(userID, profilefieldchangemodel.FieldUsername)
}

func (s *UserProfileService) GetProfiles(userIDs []uint) ([]userprofilemodel.UserProfile, error) {
	return s.profileRepo.FindByUserIDs(userIDs)
}

func (s *UserProfileService) GetUsers(userIDs []uint) ([]usermodel.User, error) {
	return s.userRepo.FindByIDs(userIDs)
}

func (s *UserProfileService) GetUserFaculties(userIDs []uint) ([]usermodel.UserFaculty, error) {
	return s.userFacultyRepo.FindByUserIDs(userIDs)
}

func (s *UserProfileService) GetURL(objectName string) string {
	return s.storageUsecase.GetURL(objectName)
}
