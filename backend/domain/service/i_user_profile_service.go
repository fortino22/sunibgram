package service

import (
	"context"
	"mime/multipart"

	"github.com/sunibgram/domain/model"
)

type FullProfile struct {
	User    *model.User
	Stat    *model.UserStat
	Profile *model.UserProfile
}

type UpsertProfileInput struct {
	Avatar *multipart.FileHeader
	Banner *multipart.FileHeader
	Bio    *string
}

type UpdateProfileInput struct {
	Name     *string
	Username *string
	Avatar   *multipart.FileHeader
	Banner   *multipart.FileHeader
	Bio      *string
}

type IUserProfileService interface {
	UpsertProfile(ctx context.Context, userID uint, input UpsertProfileInput) error
	UpdateProfile(ctx context.Context, userID uint, input UpdateProfileInput) error
	GetProfile(userID uint) (*FullProfile, error)
	GetURL(bucket, objectName string) string
}
