package repository

import (
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
	FindByCommunityID(communityID uint, page, limit int) ([]model.Thread, int64, error)
	FindFollowingFeed(userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error)
}
