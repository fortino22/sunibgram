package service

import (
	"context"
	"mime/multipart"
	"time"

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
}

type UpsertProfileInput struct {
	Avatar *multipart.FileHeader
	Banner *multipart.FileHeader
	Bio    *string
}

type UpdateProfileInput struct {
	Name         *string
	Username     *string
	Avatar       *multipart.FileHeader
	Banner       *multipart.FileHeader
	Bio          *string
	BirthdayDate *time.Time
	ClearAvatar  bool
	ClearBanner  bool
}

type IUserProfileService interface {
	UpsertProfile(ctx context.Context, userID uint, input UpsertProfileInput) error
	UpdateProfile(ctx context.Context, userID uint, input UpdateProfileInput) error
	GetProfile(userID uint, viewerID *uint) (*FullProfile, error)
	GetProfiles(userIDs []uint) ([]userprofilemodel.UserProfile, error)
	GetUsers(userIDs []uint) ([]usermodel.User, error)
	GetUserFaculties(userIDs []uint) ([]usermodel.UserFaculty, error)
	GetURL(bucket, objectName string) string
}
