package service

import (
	"context"
	"mime/multipart"

	"github.com/sunibgram/domain/model"
)

type FullProfile struct {
	User        *model.User
	Stat        *model.UserStat
	Profile     *model.UserProfile
	UserFaculty *model.UserFaculty
	IsFollowing bool
}

type UpsertProfileInput struct {
	Avatar *multipart.FileHeader
	Banner *multipart.FileHeader
	Bio    *string
}

type UpdateProfileInput struct {
	Name        *string
	Username    *string
	Avatar      *multipart.FileHeader
	Banner      *multipart.FileHeader
	Bio         *string
	ClearAvatar bool
	ClearBanner bool
}

type IUserProfileService interface {
	UpsertProfile(ctx context.Context, userID uint, input UpsertProfileInput) error
	UpdateProfile(ctx context.Context, userID uint, input UpdateProfileInput) error
	GetProfile(userID uint, viewerID *uint) (*FullProfile, error)
	GetProfiles(userIDs []uint) ([]model.UserProfile, error)
	GetUsers(userIDs []uint) ([]model.User, error)
	GetUserFaculties(userIDs []uint) ([]model.UserFaculty, error)
	GetURL(bucket, objectName string) string
}
