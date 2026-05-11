package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sunibgram/domain/model"
)

type IHashtagRepository interface {
	UpsertHashtags(ctx context.Context, names []string) (map[string]uuid.UUID, error)
	LinkHashtagsToThread(ctx context.Context, threadID uint, hashtagIDs []uuid.UUID) error
	ReplaceHashtagsForThread(ctx context.Context, threadID uint, hashtagIDs []uuid.UUID) error
	GetTrending(ctx context.Context, limit int, since24h, since7d time.Time) ([]model.TrendingHashtag, error)
	GetThreadsByHashtag(ctx context.Context, name string, cursor *uint, userID uint, limit int) ([]model.Thread, *uint, error)
}
