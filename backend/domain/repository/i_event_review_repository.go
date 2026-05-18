package repository

import "github.com/sunibgram/domain/model"

type IEventReviewRepository interface {
	Create(review *model.EventReview) error
	FindByEventIDs(eventIDs []string) (map[string][]model.EventReview, error)
}
