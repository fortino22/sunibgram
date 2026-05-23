package hashtagrepoiface

import (
	"context"

	"github.com/google/uuid"

	hashtagmodel "github.com/sunibgram/domain/model/hashtag"
)

type IHashtagRepository interface {
	UpsertHashtags(ctx context.Context, names []string) (map[string]uuid.UUID, error)
	LinkHashtagsToThread(ctx context.Context, threadID uint, hashtagIDs []uuid.UUID) error
	ReplaceHashtagsForThread(ctx context.Context, threadID uint, hashtagIDs []uuid.UUID) error
	UnlinkHashtagsFromThread(ctx context.Context, threadID uint) error
	GetTrending(ctx context.Context, limit int) ([]hashtagmodel.TrendingHashtag, error)
	PatchHashtagCounterByThreadID(ctx context.Context, threadID uint, column string, delta int) error
}
