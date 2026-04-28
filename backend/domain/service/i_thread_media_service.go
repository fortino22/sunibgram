package service

import (
	"context"
	"mime/multipart"

	"github.com/sunibgram/domain/model"
)

type IThreadMediaService interface {
	Upload(ctx context.Context, threadID, userID uint, files []*multipart.FileHeader) ([]model.ThreadMedia, error)
	GetByThreadID(threadID uint) ([]model.ThreadMedia, error)
	GetByThreadIDs(threadIDs []uint) ([]model.ThreadMedia, error)
	Delete(ctx context.Context, threadID, userID uint, mediaIDs []uint) error
	DeleteByThread(ctx context.Context, threadID uint) error
}
