package repository

import (
	"context"

	exploremodel "github.com/sunibgram/domain/model/explore"
	threadmodel "github.com/sunibgram/domain/model/thread"
)

type IExploreRepository interface {
	ExploreSearch(ctx context.Context, query string, requesterID *uint, limit int) (*exploremodel.ExploreResult, error)
	FindTopPosts(ctx context.Context, query string, requesterID uint, limit int) ([]threadmodel.Thread, int64, error)
	FindLatestPosts(ctx context.Context, query string, requesterID uint, limit int) ([]threadmodel.Thread, int64, error)
	FindPeople(ctx context.Context, query string, requesterID uint, limit int) ([]exploremodel.ExploreUserResult, error)
	FindMediaPosts(ctx context.Context, query string, requesterID uint, limit int) ([]threadmodel.Thread, int64, error)
	FindCommunities(ctx context.Context, query string, requesterID uint, limit int) ([]exploremodel.ExploreCommunityResult, int64, error)
	FindSuggestedPeople(ctx context.Context, requesterID uint, limit int) ([]exploremodel.ExploreUserResult, error)
}
