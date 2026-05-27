package explorerepoiface

import (
	"context"

	exploremodel "github.com/sunibgram/domain/model/explore"
	threadmodel "github.com/sunibgram/domain/model/thread"
)

type IExploreRepository interface {
	ExploreSearch(ctx context.Context, query string, requesterID *uint, page, limit int) (*exploremodel.ExploreResult, error)
	FindTopPosts(ctx context.Context, query string, requesterID uint, page, limit int) ([]threadmodel.Thread, int64, error)
	FindLatestPosts(ctx context.Context, query string, requesterID uint, page, limit int) ([]threadmodel.Thread, int64, error)
	FindPeople(ctx context.Context, query string, requesterID uint, page, limit int) ([]exploremodel.ExploreUserResult, error)
	FindMediaPosts(ctx context.Context, query string, requesterID uint, page, limit int) ([]threadmodel.Thread, int64, error)
	FindCommunities(ctx context.Context, query string, requesterID uint, page, limit int) ([]exploremodel.ExploreCommunityResult, int64, error)
	FindFacultySuggestions(ctx context.Context, requesterID uint, page, limit int) ([]exploremodel.ExploreUserResult, error)
	FindMutualSuggestions(ctx context.Context, requesterID uint, page, limit int) ([]exploremodel.ExploreUserResult, error)
	FindRandomSuggestions(ctx context.Context, requesterID uint, page, limit int) ([]exploremodel.ExploreUserResult, error)
}
