package service

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type IThreadService interface {
	Create(ctx context.Context, userID uint, content string, parentID *uint, eventID *string) (*model.Thread, error)
	GetByID(id uint) (*model.Thread, error)
	GetRoots(ctx context.Context, requesterID *uint, page, limit int) ([]model.Thread, int64, error)
	GetChain(ctx context.Context, rootID, requesterID uint, page, limit int) ([]model.Thread, int64, error)
	Update(ctx context.Context, id, userID uint, content string) (*model.Thread, error)
	Delete(ctx context.Context, id, userID uint) error
	Repost(ctx context.Context, userID, threadID uint) error
	Unrepost(ctx context.Context, userID, threadID uint) error
	IsReposted(userID, threadID uint) bool
	CanViewUserThreads(ctx context.Context, requesterID, targetUserID uint) error
	GetMyPosts(ctx context.Context, requesterID, userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error)
	GetMyReplies(ctx context.Context, requesterID, userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error)
	GetMyMediaThread(ctx context.Context, requesterID, userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error)
	GetMentionUsers(ctx context.Context, userID uint, query string, limit int) ([]model.ExploreUserResult, error)
	CreateCommunityThread(ctx context.Context, communityID, userID uint, content string, eventID *string) (*model.Thread, error)
	GetCommunityThreadTop(communityID, userID uint, page, limit int) ([]model.Thread, int64, error)
	GetCommunityThreadLatest(communityID, userID uint, page, limit int) ([]model.Thread, int64, error)
	GetCommunityThreadMedia(communityID, userID uint) ([]model.CommunityThreadMedia, error)
	GetFollowingFeed(userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error)
	GetFacultyThreads(facultyID uint, page, limit int) ([]model.Thread, int64, error)
}
