package feedbackmediasvciface

import (
	"context"
	"mime/multipart"

	"github.com/google/uuid"

	feedbackmediamodel "github.com/sunibgram/domain/model/feedback_media"
)

type IFeedbackMediaService interface {
	Upload(ctx context.Context, feedbackID uuid.UUID, files []*multipart.FileHeader) ([]feedbackmediamodel.FeedbackMedia, error)
	GetByFeedbackID(feedbackID uuid.UUID) ([]feedbackmediamodel.FeedbackMedia, error)
	GetByFeedbackIDs(feedbackIDs []uuid.UUID) ([]feedbackmediamodel.FeedbackMedia, error)
	DeleteByFeedback(ctx context.Context, feedbackID uuid.UUID) error
	GetURL(objectName string) string
}
