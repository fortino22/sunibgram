package usecase

import (
	"context"

	"github.com/sunibgram/application/helper"
	threadmodel "github.com/sunibgram/domain/model/thread"
)

func (s *ThreadService) filterViewableThreads(ctx context.Context, requesterID *uint, threads []threadmodel.Thread) ([]threadmodel.Thread, error) {
	return helper.FilterViewableThreads(ctx, s.settingsRepo, s.followRepo, requesterID, threads)
}
