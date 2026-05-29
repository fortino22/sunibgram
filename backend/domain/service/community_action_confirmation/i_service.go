package communityactionconfirmationsvciface

import "context"

type ICommunityActionConfirmationService interface {
	RecordAdminTransfer(ctx context.Context, communityID, fromUserID, toUserID uint) error
	RecordCommunityDeletion(ctx context.Context, communityID, actorUserID uint, communityName string) error
}
