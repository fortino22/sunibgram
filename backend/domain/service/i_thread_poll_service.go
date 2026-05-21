package service

import (
	"context"
	"mime/multipart"

	"github.com/sunibgram/domain/model"
)

type CreatePollInput struct {
	Question        string
	Labels          []string
	Images          []*multipart.FileHeader
	DurationSeconds int
}

type IThreadPollService interface {
	Create(ctx context.Context, threadID, userID uint, input CreatePollInput) (*model.ThreadPoll, error)
	GetByThreadID(threadID, requesterID uint) (*model.ThreadPoll, *uint, error)
	GetForThreads(threadIDs []uint, requesterID uint) (map[uint]*model.ThreadPoll, map[uint]uint, error)
	Vote(ctx context.Context, threadID, optionID, userID uint) (*model.ThreadPoll, *uint, error)
	DeleteByThread(ctx context.Context, threadID uint) error
	GetImageURL(bucket, objectName string) string
	HasPoll(threadID uint) bool
}
