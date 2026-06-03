package feedbackmediarepo

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	feedbackmediamodel "github.com/sunibgram/domain/model/feedback_media"
	feedbackmediarepoiface "github.com/sunibgram/domain/repository/feedback_media"
)

type FeedbackMediaRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewFeedbackMediaRepository(writeDB, readDB *gorm.DB) feedbackmediarepoiface.IFeedbackMediaRepository {
	return &FeedbackMediaRepository{writeDB: writeDB, readDB: readDB}
}

func (r *FeedbackMediaRepository) SaveAll(media []feedbackmediamodel.FeedbackMedia) error {
	if len(media) == 0 {
		return nil
	}
	return r.writeDB.Create(&media).Error
}

func (r *FeedbackMediaRepository) FindByFeedbackID(feedbackID uuid.UUID) ([]feedbackmediamodel.FeedbackMedia, error) {
	var media []feedbackmediamodel.FeedbackMedia
	if err := r.readDB.Where("feedback_id = ?", feedbackID).Order("media_id ASC").Find(&media).Error; err != nil {
		return nil, err
	}
	return media, nil
}

func (r *FeedbackMediaRepository) FindByFeedbackIDs(feedbackIDs []uuid.UUID) ([]feedbackmediamodel.FeedbackMedia, error) {
	var media []feedbackmediamodel.FeedbackMedia
	if len(feedbackIDs) == 0 {
		return media, nil
	}
	if err := r.readDB.Where("feedback_id IN ?", feedbackIDs).Order("media_id ASC").Find(&media).Error; err != nil {
		return nil, err
	}
	return media, nil
}

func (r *FeedbackMediaRepository) DeleteByFeedbackID(feedbackID uuid.UUID) error {
	return r.writeDB.Where("feedback_id = ?", feedbackID).Delete(&feedbackmediamodel.FeedbackMedia{}).Error
}
