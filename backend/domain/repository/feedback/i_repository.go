package feedbackrepoiface

import (
	"github.com/google/uuid"

	feedbackmodel "github.com/sunibgram/domain/model/feedback"
)

type ListFilter struct {
	Type   string
	Status string
	Page   int
	Limit  int
}

type IFeedbackRepository interface {
	Save(f *feedbackmodel.Feedback) error
	FindByID(id uuid.UUID) (*feedbackmodel.Feedback, error)
	FindAll(filter ListFilter) ([]feedbackmodel.Feedback, int64, error)
	Update(f *feedbackmodel.Feedback) error
	UpdateStatus(id uuid.UUID, status string) error
	Delete(id uuid.UUID) error

	UpsertVote(feedbackID uuid.UUID, userID uint, value int) error
	DeleteVote(feedbackID uuid.UUID, userID uint) error
	GetVoteCounts(feedbackID uuid.UUID) (feedbackmodel.VoteCounts, error)
	GetUserVote(feedbackID uuid.UUID, userID uint) (int, bool, error)
	GetUserVotes(feedbackIDs []uuid.UUID, userID uint) (map[uuid.UUID]int, error)
	GetVoteCountsBulk(feedbackIDs []uuid.UUID) (map[uuid.UUID]feedbackmodel.VoteCounts, error)
}
