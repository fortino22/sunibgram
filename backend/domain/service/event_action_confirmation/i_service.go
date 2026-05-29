package eventactionconfirmationsvciface

import "context"

type IEventActionConfirmationService interface {
	RecordEventDeletion(ctx context.Context, eventID string, actorUserID uint, eventTitle string) error
}
