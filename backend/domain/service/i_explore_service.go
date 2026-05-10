package service

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type IExploreService interface {
	ExploreSearch(ctx context.Context, query string, requesterID *uint) (*model.ExploreResult, error)
	GetTopPosts(ctx context.Context, query string, requesterID uint) ([]model.Thread, int64, error)
	GetLatestPosts(ctx context.Context, query string, requesterID uint) ([]model.Thread, int64, error)
	GetPeople(ctx context.Context, query string, requesterID uint) ([]model.ExploreUserResult, error)
	GetMediaPosts(ctx context.Context, query string, requesterID uint) ([]model.Thread, int64, error)
	GetCommunities(ctx context.Context, query string, requesterID uint) ([]model.ExploreCommunityResult, int64, error)
}
