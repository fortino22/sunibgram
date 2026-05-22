package repository

import (
	"gorm.io/gorm"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
)

type eventReviewRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewEventReviewRepository(writeDB, readDB *gorm.DB) domainrepo.IEventReviewRepository {
	return &eventReviewRepository{writeDB: writeDB, readDB: readDB}
}

func (r *eventReviewRepository) Create(review *model.EventReview) error {
	return r.writeDB.Create(review).Error
}

func (r *eventReviewRepository) FindByEventIDs(eventIDs []string) (map[string][]model.EventReview, error) {
	reviewMap := make(map[string][]model.EventReview)
	if len(eventIDs) == 0 {
		return reviewMap, nil
	}

	var reviews []model.EventReview
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
