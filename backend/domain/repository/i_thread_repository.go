package repository

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type IThreadRepository interface {
	Save(thread *model.Thread) error
	InvalidateThread(threadID uint)
	FindByID(id uint) (*model.Thread, error)
	FindRoots(page, limit int) ([]model.Thread, int64, error)
	FindByRootID(rootID uint, page, limit int) ([]model.Thread, int64, error)
	Update(thread *model.Thread) error
	Delete(id uint) error
	FindPostsByUserID(requesterID, userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error)
	FindReplyTabByUserID(requesterID, userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error)
	FindMediaByUserID(requesterID, userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error)
	ExistsRepost(userID, originalThreadID uint) (bool, error)
	DeleteRepost(userID, originalThreadID uint) error
	FindCommunityTop(communityID uint, page, limit int) ([]model.Thread, int64, error)
	FindCommunityLatest(communityID uint, page, limit int) ([]model.Thread, int64, error)
	FindCommunityMedia(communityID uint) ([]model.CommunityThreadMedia, error)
	FindFollowingFeed(userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error)
	FindForYouFeed(ctx context.Context, userID uint, facultyID uint, page, limit int) ([]model.Thread, int64, error)
	FindByFacultyID(facultyID uint, page, limit int) ([]model.Thread, int64, error)
}
