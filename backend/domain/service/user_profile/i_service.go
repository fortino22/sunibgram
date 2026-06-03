package userprofilesvciface

import (
	"context"
	"mime/multipart"
	"time"

	profilefieldchangemodel "github.com/sunibgram/domain/model/profile_field_change"
	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
	userstatmodel "github.com/sunibgram/domain/model/user_stat"
)

type FullProfile struct {
	User        *usermodel.User
	Stat        *userstatmodel.UserStat
	Profile     *userprofilemodel.UserProfile
	UserFaculty *usermodel.UserFaculty
	IsFollowing bool
	IsBlocked   bool
	IsBlockedBy bool
	IsMuted     bool
	// *ChangeAvailableAt are pre-computed earliest-next-change times for the
	// rate-limited profile fields, set only while still within the cooldown.
	// Rendered self-view only by the controller.
	NameChangeAvailableAt            *time.Time
	UsernameChangeAvailableAt        *time.Time
	FacultyLocationChangeAvailableAt *time.Time
}

type UpsertProfileInput struct {
	Avatar *multipart.FileHeader
	Banner *multipart.FileHeader
	Bio    *string
}

type UpdateProfileInput struct {
	Name              *string
	Username          *string
	Avatar            *multipart.FileHeader
	Banner            *multipart.FileHeader
	Bio               *string
	BirthdayDate      *time.Time
	FacultyLocationID *uint
	ClearAvatar       bool
	ClearBanner       bool
}

type IUserProfileService interface {
	UpsertProfile(ctx context.Context, userID uint, input UpsertProfileInput) error
	UpdateProfile(ctx context.Context, userID uint, input UpdateProfileInput) error
	GetProfile(userID uint, viewerID *uint) (*FullProfile, error)
	// UsernameHistory returns the user's username-change log, newest first.
	UsernameHistory(userID uint) ([]profilefieldchangemodel.ProfileFieldChange, error)
	CompleteOnboarding(userID uint) error
	// OnboardingStatus reports the dynamic follow requirement (required) and the
	// user's current following count for the first-time onboarding gate.
	OnboardingStatus(userID uint) (required int, following int, err error)
	GetProfiles(userIDs []uint) ([]userprofilemodel.UserProfile, error)
	GetUsers(userIDs []uint) ([]usermodel.User, error)
	GetUserFaculties(userIDs []uint) ([]usermodel.UserFaculty, error)
	GetURL(objectName string) string
}
