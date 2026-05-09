package service

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type IExploreService interface {
	ExploreSearch(ctx context.Context, query string, requesterID *uint) (*model.ExploreResult, error)
}
