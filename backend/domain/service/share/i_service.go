package sharesvciface

import (
	"context"

	exploremodel "github.com/sunibgram/domain/model/explore"
)

type IShareService interface {
	RecordCopy(ctx context.Context, userID, threadID uint) (int, error)
	SendToUsers(ctx context.Context, senderID, threadID uint, recipientIDs []uint, message *string) (int, error)
	SearchUsers(ctx context.Context, query string, requesterID uint) ([]exploremodel.ExploreUserResult, error)
}
