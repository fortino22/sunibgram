package service

import (
	"context"
	"mime/multipart"

	"github.com/sunibgram/domain/model"
)

type IAvatarService interface {
	UploadAvatar(ctx context.Context, userID uint, file multipart.File, header *multipart.FileHeader) (*model.UserAvatar, error)
	GetAvatarByUserID(userID uint) (*model.UserAvatar, error)
	DeleteAvatarByUserID(ctx context.Context, userID uint) error
}
