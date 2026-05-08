package repository

import (
	"context"
	"fmt"
)

func threadKey(id uint) string {
	return fmt.Sprintf("thread:id:%d", id)
}

func threadRootsKey(page, limit int) string {
	return fmt.Sprintf("thread:roots:page:%d:limit:%d", page, limit)
}

func threadChainKey(rootID uint, page, limit int) string {
	return fmt.Sprintf("thread:chain:%d:page:%d:limit:%d", rootID, page, limit)
}

func threadCommunityKey(communityID uint, page, limit int) string {
	return fmt.Sprintf("thread:community:%d:page:%d:limit:%d", communityID, page, limit)
}

func (r *threadCached) invalidate(ctx context.Context, threadID uint, rootID *uint, communityID *uint) {
	r.cache.Delete(ctx, threadKey(threadID))

	treeID := threadID
	if rootID != nil && *rootID != 0 {
		treeID = *rootID
	}
	r.cache.DeleteByPattern(ctx, fmt.Sprintf("thread:chain:%d:*", treeID))

	if communityID != nil && *communityID != 0 {
		r.cache.DeleteByPattern(ctx, fmt.Sprintf("thread:community:%d:*", *communityID))
	}
}

func (r *threadCached) invalidateInteractions(ctx context.Context, threadID uint) {
	r.likes.invalidateByThreadID(ctx, threadID)
	r.bookmarks.invalidateByThreadID(ctx, threadID)
	r.reposts.invalidateByThreadID(ctx, threadID)
}
