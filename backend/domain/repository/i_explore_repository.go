package repository

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type IExploreRepository interface {
	ExploreSearch(ctx context.Context, query string, requesterID *uint, limit int) (*model.ExploreResult, error)
	FindTopPosts(ctx context.Context, query string, limit int) ([]model.Thread, int64, error)
	FindLatestPosts(ctx context.Context, query string, limit int) ([]model.Thread, int64, error)
	FindPeople(ctx context.Context, query string, requesterID uint, limit int) ([]model.ExploreUserResult, error)
	FindMediaPosts(ctx context.Context, query string, limit int) ([]model.Thread, int64, error)
	FindCommunities(ctx context.Context, query string, requesterID uint, limit int) ([]model.ExploreCommunityResult, int64, error)
	FindSuggestedPeople(ctx context.Context, requesterID uint, limit int) ([]model.ExploreUserResult, error)
}
