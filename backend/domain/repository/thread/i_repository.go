package threadrepoiface

import (
	"context"

	"github.com/google/uuid"

	threadmodel "github.com/sunibgram/domain/model/thread"
	threadmediamodel "github.com/sunibgram/domain/model/thread_media"
)

type IThreadRepository interface {
	Save(thread *threadmodel.Thread) error
	InvalidateThread(threadID uint)
	FindByID(id uint) (*threadmodel.Thread, error)
	FindRoots(requesterID *uint, page, limit int) ([]threadmodel.Thread, int64, error)
	FindByRootID(rootID uint, page, limit int) ([]threadmodel.Thread, int64, error)
	FindChain(anchorID uint) (threadmodel.Thread, []threadmodel.Thread, []threadmodel.Thread, error)
	Update(thread *threadmodel.Thread) error
	UpdateReplyPermission(threadID uint, replyPermission threadmodel.ReplyPermission) error
	Delete(id uint) error
	FindPostsByUserID(requesterID, userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	FindReplyTabByUserID(requesterID, userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	FindMediaByUserID(requesterID, userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	ExistsRepost(userID, originalThreadID uint) (bool, error)
	AreThreadsReposted(userID uint, threadIDs []uint) (map[uint]bool, error)
	AllRepostedOriginalIDs(userID uint) ([]uint, error)
	RepostWrapperIDsByOriginal(originalThreadID uint) ([]uint, error)
	DeleteRepost(userID, originalThreadID uint) error
	FindCommunityTop(communityID, requesterUserID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	FindCommunityLatest(communityID, requesterUserID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	FindCommunityThreadsBySort(communityID, requesterUserID uint, sort string, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	FindCommunityMedia(communityID, requesterUserID uint) ([]threadmediamodel.CommunityThreadMedia, error)
	FindMyCommunityThreads(userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	FindAllCommunityThreads(userID uint, categoryID *uuid.UUID, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	FindAllCommunityThreadsByQueryTop(userID uint, query string, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	FindAllCommunityThreadsByQueryLatest(userID uint, query string, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	FindFollowingFeed(userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	FindForYouFeed(ctx context.Context, userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	InvalidateForYouSeed(ctx context.Context, userID uint)
	CountNewSinceID(ctx context.Context, userID, sinceID uint) (int64, error)
	FindNewSinceID(ctx context.Context, userID, sinceID uint, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, error)
	FindByFacultyID(facultyID uint, requesterID *uint, page, limit int) ([]threadmodel.Thread, int64, error)
	AttachReplyParents(threads []threadmodel.Thread, requesterID *uint) ([]threadmodel.Thread, []*threadmodel.Thread, error)
	InvalidateUserFeedCaches(ctx context.Context, userID uint)
	FindRepliesContent(parentID uint, limit int) ([]string, error)
	IncrementShareCount(threadID uint) error
}
