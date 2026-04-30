package service

import "github.com/sunibgram/domain/model"

type IFollowService interface {
	Follow(followerID, followingID uint) error
	Unfollow(followerID, followingID uint) error
	GetStat(userID uint) (*model.UserStat, error)
}
