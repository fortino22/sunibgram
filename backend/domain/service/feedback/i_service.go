package feedbacksvciface

import (
	"github.com/google/uuid"

	feedbackmodel "github.com/sunibgram/domain/model/feedback"
)

type FeedbackView struct {
	Feedback feedbackmodel.Feedback
	Counts   feedbackmodel.VoteCounts
	UserVote int // -1, 0, 1
}

type IFeedbackService interface {
	Create(userID uint, fType, title, body string) error
	GetByID(viewerID uint, id uuid.UUID) (*FeedbackView, error)
	List(viewerID uint, fType, status string, page, limit int) ([]FeedbackView, int64, error)
	UpdateStatus(adminID uint, id uuid.UUID, status string) error
	Delete(userID uint, roleName string, id uuid.UUID) error

	Vote(userID uint, id uuid.UUID, value int) error
}
