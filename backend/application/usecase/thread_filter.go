package usecase

import (
	"context"

	"github.com/sunibgram/domain/model"
)

func (s *ThreadService) filterViewableThreads(ctx context.Context, requesterID *uint, threads []model.Thread) ([]model.Thread, error) {
	targetIDs := make([]uint, 0, len(threads))
	seen := make(map[uint]struct{}, len(threads))
	for _, thread := range threads {
		if requesterID != nil && thread.UserID == *requesterID {
			continue
		}
		if _, ok := seen[thread.UserID]; ok {
			continue
		}
		seen[thread.UserID] = struct{}{}
		targetIDs = append(targetIDs, thread.UserID)
	}

	privacyMap, err := s.settingsRepo.GetPrivacyMapByUserIDs(ctx, targetIDs)
	if err != nil {
		return nil, err
	}

	privateIDs := make([]uint, 0, len(privacyMap))
	for _, id := range targetIDs {
		if privacyMap[id] {
			privateIDs = append(privateIDs, id)
		}
	}

	followingSet := map[uint]bool{}
	if requesterID != nil {
		followingSet, err = s.followRepo.GetFollowingSet(ctx, *requesterID, privateIDs)
		if err != nil {
			return nil, err
		}
	}

	filtered := make([]model.Thread, 0, len(threads))
	for _, thread := range threads {
		if requesterID != nil && thread.UserID == *requesterID {
			filtered = append(filtered, thread)
			continue
		}
		if !privacyMap[thread.UserID] || followingSet[thread.UserID] {
			filtered = append(filtered, thread)
		}
	}
	return filtered, nil
}
