package helper

import (
	"context"

	threadmodel "github.com/sunibgram/domain/model/thread"
	followrepoiface "github.com/sunibgram/domain/repository/follow"
	usersettingsrepoiface "github.com/sunibgram/domain/repository/user_settings"
)

func FilterViewableThreads(ctx context.Context, settingsRepo usersettingsrepoiface.IUserSettingsRepository, followRepo followrepoiface.IFollowRepository, requesterID *uint, threads []threadmodel.Thread) ([]threadmodel.Thread, error) {
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

	privacyMap, err := settingsRepo.GetPrivacyMapByUserIDs(ctx, targetIDs)
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
		followingSet, err = followRepo.GetFollowingSet(ctx, *requesterID, privateIDs)
		if err != nil {
			return nil, err
		}
	}

	filtered := make([]threadmodel.Thread, 0, len(threads))
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
