package exploresvciface

import (
	"context"

	exploremodel "github.com/sunibgram/domain/model/explore"
	threadmodel "github.com/sunibgram/domain/model/thread"
)

type IExploreService interface {
	ExploreSearch(ctx context.Context, query string, requesterID *uint, page, limit int) (*exploremodel.ExploreResult, error)
	GetTopPosts(ctx context.Context, query string, requesterID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	GetLatestPosts(ctx context.Context, query string, requesterID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	GetPeople(ctx context.Context, query string, requesterID uint, page, limit int) ([]exploremodel.ExploreUserResult, error)
	GetMediaPosts(ctx context.Context, query string, requesterID uint, page, limit int) ([]threadmodel.Thread, int64, error)
	GetCommunities(ctx context.Context, query string, requesterID uint, page, limit int) ([]exploremodel.ExploreCommunityResult, int64, error)
	GetFacultySuggestions(ctx context.Context, requesterID uint, page, limit int) ([]exploremodel.ExploreUserResult, error)
	GetMutualSuggestions(ctx context.Context, requesterID uint, page, limit int) ([]exploremodel.ExploreUserResult, error)
	GetSuggestedPeoplePreview(ctx context.Context, requesterID uint) ([]exploremodel.ExploreUserResult, error)
}
