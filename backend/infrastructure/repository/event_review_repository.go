package repository

import (
	"gorm.io/gorm"

	eventreviewmodel "github.com/sunibgram/domain/model/event_review"
	domainrepo "github.com/sunibgram/domain/repository"
)

type eventReviewRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewEventReviewRepository(writeDB, readDB *gorm.DB) domainrepo.IEventReviewRepository {
	return &eventReviewRepository{writeDB: writeDB, readDB: readDB}
}

func (r *eventReviewRepository) Create(review *eventreviewmodel.EventReview) error {
	return r.writeDB.Create(review).Error
}

func (r *eventReviewRepository) FindByEventIDs(eventIDs []string) (map[string][]eventreviewmodel.EventReview, error) {
	reviewMap := make(map[string][]eventreviewmodel.EventReview)
	if len(eventIDs) == 0 {
		return reviewMap, nil
	}

	var reviews []eventreviewmodel.EventReview
	err := r.readDB.
		Preload("Admin").
		Where("event_id IN ?", eventIDs).
		Where("action IN ?", []string{"declined", "revision"}).
		Order("event_id ASC, created_at ASC").
		Find(&reviews).Error
	if err != nil {
		return nil, err
	}
	for _, review := range reviews {
		reviewMap[review.EventID] = append(reviewMap[review.EventID], review)
	}
	return reviewMap, nil
}
