package repository

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type IFollowRepository interface {
	Follow(followerID, followingID uint) error
	Unfollow(followerID, followingID uint) error
	Exists(followerID, followingID uint) (bool, error)
	GetFollowingSet(ctx context.Context, followerID uint, targetIDs []uint) (map[uint]bool, error)
	GetStat(userID uint) (*model.UserStat, error)
	GetFollowing(userID uint, page, limit int) ([]model.User, int64, error)
	GetFollowers(userID uint, page, limit int) ([]model.User, int64, error)
}
