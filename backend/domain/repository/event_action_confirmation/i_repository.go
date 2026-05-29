package eventactionconfirmationrepoiface

import (
	eventactionconfirmationmodel "github.com/sunibgram/domain/model/event_action_confirmation"
)

type IEventActionConfirmationRepository interface {
	Save(rec *eventactionconfirmationmodel.EventActionConfirmation) error
}
