package service

import (
	"context"
	"mime/multipart"

	"github.com/sunibgram/domain/model"
)

type ICommunityService interface {
	Create(ctx context.Context, userID uint, name, description, rules, category string, visibility model.CommunityVisibility, avatar, banner *multipart.FileHeader) (*model.Community, error)
	Update(ctx context.Context, id, userID uint, name, description, rules, category string, visibility model.CommunityVisibility, avatar, banner *multipart.FileHeader, clearAvatar, clearBanner bool) (*model.Community, error)
	Delete(ctx context.Context, id, userID uint) error
	GetByID(id uint) (*model.Community, error)
	GetAll(page, limit int, category string) ([]model.Community, int64, error)
	GetMyCreated(userID uint, page, limit int) ([]model.Community, int64, error)
	GetMyJoined(userID uint, page, limit int) ([]model.Community, int64, error)
	GetURL(bucket, objectName string) string
}
