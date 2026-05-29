package eventactionconfirmationrepo

import (
	"gorm.io/gorm"

	eventactionconfirmationmodel "github.com/sunibgram/domain/model/event_action_confirmation"
	eventactionconfirmationrepoiface "github.com/sunibgram/domain/repository/event_action_confirmation"
)

type eventActionConfirmationRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewEventActionConfirmationRepository(writeDB, readDB *gorm.DB) eventactionconfirmationrepoiface.IEventActionConfirmationRepository {
	return &eventActionConfirmationRepository{writeDB: writeDB, readDB: readDB}
}

func (r *eventActionConfirmationRepository) Save(rec *eventactionconfirmationmodel.EventActionConfirmation) error {
	return r.writeDB.Create(rec).Error
}
