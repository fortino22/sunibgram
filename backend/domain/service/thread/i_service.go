package threadsvciface

import (
	"context"

	"github.com/google/uuid"

	exploremodel "github.com/sunibgram/domain/model/explore"
	threadmodel "github.com/sunibgram/domain/model/thread"
	threadmediamodel "github.com/sunibgram/domain/model/thread_media"
)

type IThreadService interface {
	Create(ctx context.Context, userID uint, content string, parentID *uint, eventID *string, replyPermission *string) (*threadmodel.Thread, error)
	CreateCommunityThread(ctx context.Context, communityID, userID uint, content string, eventID *string) (*threadmodel.Thread, error)
	UpdateReplyPermission(ctx context.Context, threadID, userID uint, replyPermission string) error
	GetByID(id uint) (*threadmodel.Thread, error)
	GetRoots(ctx context.Context, requesterID *uint, page, limit int) ([]threadmodel.Thread, int64, error)
	GetChain(ctx context.Context, rootID, requesterID uint, page, limit int) ([]threadmodel.Thread, int64, error)
	Update(ctx context.Context, id, userID uint, content string) (*threadmodel.Thread, error)
	Delete(ctx context.Context, id, userID uint) error
	Repost(ctx context.Context, userID, threadID uint) error
	Unrepost(ctx context.Context, userID, threadID uint) error
	IsReposted(userID, threadID uint) bool
	AreThreadsReposted(userID uint, threadIDs []uint) (map[uint]bool, error)
	CanViewUserThreads(ctx context.Context, requesterID, targetUserID uint) error
	GetMyPosts(ctx context.Context, requesterID, userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	GetMyReplies(ctx context.Context, requesterID, userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	GetMyMediaThread(ctx context.Context, requesterID, userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	GetMentionUsers(ctx context.Context, userID uint, query string, limit int) ([]exploremodel.ExploreUserResult, error)
	GetMyCommunityThreads(ctx context.Context, userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	GetAllCommunityThreads(ctx context.Context, userID uint, categoryID *uuid.UUID, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	GetAllCommunityThreadsByQueryTop(ctx context.Context, userID uint, query string, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	GetAllCommunityThreadsByQueryLatest(ctx context.Context, userID uint, query string, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	GetCommunityThreadTop(communityID, userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	GetCommunityThreadLatest(communityID, userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	GetCommunityThreadsBySort(communityID, userID uint, sort string, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	GetCommunityThreadMedia(communityID, userID uint) ([]threadmediamodel.CommunityThreadMedia, error)
	GetFollowingFeed(userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	GetForYouFeed(ctx context.Context, userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	RefreshForYouFeed(ctx context.Context, userID uint)
	CountNewForYouSinceID(ctx context.Context, userID, sinceID uint) (int64, error)
	GetFacultyThreads(facultyID, requesterID uint, page, limit int) ([]threadmodel.Thread, int64, error)
}
