package service

import (
	"context"
	"mime/multipart"

	threadmediamodel "github.com/sunibgram/domain/model/thread_media"
)

type IThreadMediaService interface {
	Upload(ctx context.Context, threadID, userID uint, files []*multipart.FileHeader) ([]threadmediamodel.ThreadMedia, error)
	GetByThreadID(threadID uint) ([]threadmediamodel.ThreadMedia, error)
	GetByThreadIDs(threadIDs []uint) ([]threadmediamodel.ThreadMedia, error)
	Update(ctx context.Context, threadID, userID uint, files []*multipart.FileHeader) error
	UpdateGranular(ctx context.Context, threadID, userID uint, keepIDs []uint, files []*multipart.FileHeader) error
	Delete(ctx context.Context, threadID, userID uint, mediaIDs []uint) error
	DeleteByThread(ctx context.Context, threadID uint) error
	GetURL(bucket, objectName string) string
}
