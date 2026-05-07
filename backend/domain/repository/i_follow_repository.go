package repository

import "github.com/sunibgram/domain/model"

type IFollowRepository interface {
	Follow(followerID, followingID uint) error
	Unfollow(followerID, followingID uint) error
	Exists(followerID, followingID uint) (bool, error)
	GetStat(userID uint) (*model.UserStat, error)
	GetFollowing(userID uint, page, limit int) ([]model.User, int64, error)
	GetFollowers(userID uint, page, limit int) ([]model.User, int64, error)
}
