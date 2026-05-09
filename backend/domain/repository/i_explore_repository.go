package repository

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type IExploreRepository interface {
	ExploreSearch(ctx context.Context, query string, requesterID *uint, limit int) (*model.ExploreResult, error)
}
