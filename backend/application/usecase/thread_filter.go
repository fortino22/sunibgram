package usecase

import (
	"context"

	"github.com/sunibgram/application/helper"
	"github.com/sunibgram/domain/model"
)

func (s *ThreadService) filterViewableThreads(ctx context.Context, requesterID *uint, threads []model.Thread) ([]model.Thread, error) {
	return helper.FilterViewableThreads(ctx, s.settingsRepo, s.followRepo, requesterID, threads)
}
