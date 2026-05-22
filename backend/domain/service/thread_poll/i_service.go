package threadpollsvciface

import (
	"context"
	"mime/multipart"

	threadpollmodel "github.com/sunibgram/domain/model/thread_poll"
)

type CreatePollInput struct {
	Question        string
	Labels          []string
	Images          []*multipart.FileHeader
	DurationSeconds int
}

type IThreadPollService interface {
	Create(ctx context.Context, threadID, userID uint, input CreatePollInput) (*threadpollmodel.ThreadPoll, error)
	GetByThreadID(threadID, requesterID uint) (*threadpollmodel.ThreadPoll, *uint, error)
	GetForThreads(threadIDs []uint, requesterID uint) (map[uint]*threadpollmodel.ThreadPoll, map[uint]uint, error)
	Vote(ctx context.Context, threadID, optionID, userID uint) (*threadpollmodel.ThreadPoll, *uint, error)
	DeleteByThread(ctx context.Context, threadID uint) error
	GetImageURL(bucket, objectName string) string
	HasPoll(threadID uint) bool
}
