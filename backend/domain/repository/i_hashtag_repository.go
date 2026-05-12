package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sunibgram/domain/model"
)

type IHashtagRepository interface {
	UpsertHashtags(ctx context.Context, names []string) (map[string]uuid.UUID, error)
	LinkHashtagsToThread(ctx context.Context, threadID uint, hashtagIDs []uuid.UUID) error
	ReplaceHashtagsForThread(ctx context.Context, threadID uint, hashtagIDs []uuid.UUID) error
	UnlinkHashtagsFromThread(ctx context.Context, threadID uint) error
	GetTrending(ctx context.Context, limit int) ([]model.TrendingHashtag, error)
}
