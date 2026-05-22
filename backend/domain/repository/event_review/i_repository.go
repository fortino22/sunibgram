package eventreviewrepoiface

import eventreviewmodel "github.com/sunibgram/domain/model/event_review"

type IEventReviewRepository interface {
	Create(review *eventreviewmodel.EventReview) error
	FindByEventIDs(eventIDs []string) (map[string][]eventreviewmodel.EventReview, error)
}
