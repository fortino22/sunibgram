package threaduc

import (
	"context"

	"github.com/sunibgram/application/helper"
	threadmodel "github.com/sunibgram/domain/model/thread"
)

func (s *ThreadService) filterViewableThreads(ctx context.Context, requesterID *uint, threads []threadmodel.Thread) ([]threadmodel.Thread, error) {
	viewable, err := helper.FilterViewableThreads(ctx, s.settingsRepo, s.followRepo, requesterID, threads)
	if err != nil {
		return nil, err
	}
	blockFiltered, err := s.filterBlockedThreads(requesterID, viewable)
	if err != nil {
		return nil, err
	}
	return s.filterMutedThreads(requesterID, blockFiltered)
}

// filterMutedThreads drops threads authored by users the requester has muted
// (one-way). Mute flags are cached per pair, so the per-author lookups over a
// single chain are cheap.
func (s *ThreadService) filterMutedThreads(requesterID *uint, threads []threadmodel.Thread) ([]threadmodel.Thread, error) {
	if s.muteRepo == nil || requesterID == nil || *requesterID == 0 {
		return threads, nil
	}
	mutedByAuthor := make(map[uint]bool)
	filtered := make([]threadmodel.Thread, 0, len(threads))
	for _, thread := range threads {
		if thread.UserID == *requesterID {
			filtered = append(filtered, thread)
			continue
		}
		muted, seen := mutedByAuthor[thread.UserID]
		if !seen {
			var err error
			muted, err = s.muteRepo.IsMuted(*requesterID, thread.UserID)
			if err != nil {
				return nil, err
			}
			mutedByAuthor[thread.UserID] = muted
		}
		if !muted {
			filtered = append(filtered, thread)
		}
	}
	return filtered, nil
}

// filterBlockedThreads drops threads authored by users in a mutual block
// relationship with the requester. Block flags are cached per pair, so the
// per-author lookups over a single chain are cheap.
func (s *ThreadService) filterBlockedThreads(requesterID *uint, threads []threadmodel.Thread) ([]threadmodel.Thread, error) {
	if s.blockRepo == nil || requesterID == nil || *requesterID == 0 {
		return threads, nil
	}
	blockedByAuthor := make(map[uint]bool)
	filtered := make([]threadmodel.Thread, 0, len(threads))
	for _, thread := range threads {
		if thread.UserID == *requesterID {
			filtered = append(filtered, thread)
			continue
		}
		blocked, seen := blockedByAuthor[thread.UserID]
		if !seen {
			var err error
			blocked, err = s.blockRepo.IsBlocked(*requesterID, thread.UserID)
			if err != nil {
				return nil, err
			}
			blockedByAuthor[thread.UserID] = blocked
		}
		if !blocked {
			filtered = append(filtered, thread)
		}
	}
	return filtered, nil
}

// filterBlockedFeed drops blocked-author threads from a (threads, parents)
// feed pair, keeping the two slices index-aligned. parents[i] is the parent
// of threads[i] (may be nil). Used by feeds that bypass ApplyThreadVisibility
// (e.g. For You, whose source queries are unfiltered).
func (s *ThreadService) filterBlockedFeed(requesterID *uint, threads []threadmodel.Thread, parents []*threadmodel.Thread) ([]threadmodel.Thread, []*threadmodel.Thread, error) {
	if s.blockRepo == nil || requesterID == nil || *requesterID == 0 {
		return threads, parents, nil
	}
	blockedByAuthor := make(map[uint]bool)
	keptThreads := make([]threadmodel.Thread, 0, len(threads))
	keptParents := make([]*threadmodel.Thread, 0, len(parents))
	for i, thread := range threads {
		var parent *threadmodel.Thread
		if i < len(parents) {
			parent = parents[i]
		}
		if thread.UserID != *requesterID {
			blocked, seen := blockedByAuthor[thread.UserID]
			if !seen {
				var err error
				blocked, err = s.blockRepo.IsBlocked(*requesterID, thread.UserID)
				if err != nil {
					return nil, nil, err
				}
				blockedByAuthor[thread.UserID] = blocked
			}
			if blocked {
				continue
			}
		}
		keptThreads = append(keptThreads, thread)
		keptParents = append(keptParents, parent)
	}
	return keptThreads, keptParents, nil
}

// filterMutedFeed drops muted-author threads from a (threads, parents) feed pair
// (one-way), keeping the two slices index-aligned. Mirrors filterBlockedFeed for
// the mute relationship.
func (s *ThreadService) filterMutedFeed(requesterID *uint, threads []threadmodel.Thread, parents []*threadmodel.Thread) ([]threadmodel.Thread, []*threadmodel.Thread, error) {
	if s.muteRepo == nil || requesterID == nil || *requesterID == 0 {
		return threads, parents, nil
	}
	mutedByAuthor := make(map[uint]bool)
	keptThreads := make([]threadmodel.Thread, 0, len(threads))
	keptParents := make([]*threadmodel.Thread, 0, len(parents))
	for i, thread := range threads {
		var parent *threadmodel.Thread
		if i < len(parents) {
			parent = parents[i]
		}
		if thread.UserID != *requesterID {
			muted, seen := mutedByAuthor[thread.UserID]
			if !seen {
				var err error
				muted, err = s.muteRepo.IsMuted(*requesterID, thread.UserID)
				if err != nil {
					return nil, nil, err
				}
				mutedByAuthor[thread.UserID] = muted
			}
			if muted {
				continue
			}
		}
		keptThreads = append(keptThreads, thread)
		keptParents = append(keptParents, parent)
	}
	return keptThreads, keptParents, nil
}
