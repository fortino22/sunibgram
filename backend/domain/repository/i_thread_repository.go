package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sunibgram/domain/model"
)

type IThreadRepository interface {
	Save(thread *model.Thread) error
	InvalidateThread(threadID uint)
	FindByID(id uint) (*model.Thread, error)
	FindRoots(requesterID *uint, page, limit int) ([]model.Thread, int64, error)
	FindByRootID(rootID uint, page, limit int) ([]model.Thread, int64, error)
	FindChain(anchorID uint) (model.Thread, []model.Thread, error)
	Update(thread *model.Thread) error
	Delete(id uint) error
	FindPostsByUserID(requesterID, userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error)
	FindReplyTabByUserID(requesterID, userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error)
	FindMediaByUserID(requesterID, userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error)
	ExistsRepost(userID, originalThreadID uint) (bool, error)
	AreThreadsReposted(userID uint, threadIDs []uint) (map[uint]bool, error)
	DeleteRepost(userID, originalThreadID uint) error
	FindCommunityTop(communityID, requesterUserID uint, page, limit int) ([]model.Thread, int64, error)
	FindCommunityLatest(communityID, requesterUserID uint, page, limit int) ([]model.Thread, int64, error)
	FindCommunityMedia(communityID, requesterUserID uint) ([]model.CommunityThreadMedia, error)
	FindMyCommunityThreads(userID uint, page, limit int) ([]model.Thread, int64, error)
	FindAllCommunityThreads(userID uint, categoryID *uuid.UUID, page, limit int) ([]model.Thread, []*model.Thread, int64, error)
	FindAllCommunityThreadsByQueryTop(userID uint, query string, page, limit int) ([]model.Thread, []*model.Thread, int64, error)
	FindAllCommunityThreadsByQueryLatest(userID uint, query string, page, limit int) ([]model.Thread, []*model.Thread, int64, error)
	FindFollowingFeed(userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error)
	FindForYouFeed(ctx context.Context, userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error)
	FindByFacultyID(facultyID uint, page, limit int) ([]model.Thread, int64, error)
	AttachReplyParents(threads []model.Thread, requesterID *uint) ([]model.Thread, []*model.Thread, error)
	InvalidateUserFeedCaches(ctx context.Context, userID uint)
}
