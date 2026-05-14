package repository

import "github.com/sunibgram/domain/model"

type IEventReviewRepository interface {
	Create(review *model.EventReview) error
	FindByEventID(eventID string) ([]model.EventReview, error)
}
