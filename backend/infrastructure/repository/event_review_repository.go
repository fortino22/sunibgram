package repository

import (
	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"gorm.io/gorm"
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

func (r *eventReviewRepository) FindByEventID(eventID string) ([]model.EventReview, error) {
	var reviews []model.EventReview
	err := r.readDB.
		Where("event_id = ?", eventID).
		Order("created_at DESC").
		Find(&reviews).Error
	return reviews, err
}
