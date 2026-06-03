package feedbackmediarepoiface

import (
	"github.com/google/uuid"

	feedbackmediamodel "github.com/sunibgram/domain/model/feedback_media"
)

type IFeedbackMediaRepository interface {
	SaveAll(media []feedbackmediamodel.FeedbackMedia) error
	FindByFeedbackID(feedbackID uuid.UUID) ([]feedbackmediamodel.FeedbackMedia, error)
	FindByFeedbackIDs(feedbackIDs []uuid.UUID) ([]feedbackmediamodel.FeedbackMedia, error)
	DeleteByFeedbackID(feedbackID uuid.UUID) error
}
