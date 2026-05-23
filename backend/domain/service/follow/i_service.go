package followsvciface

import (
	usermodel "github.com/sunibgram/domain/model/user"
	userstatmodel "github.com/sunibgram/domain/model/user_stat"
)

type IFollowService interface {
	Follow(followerID, followingID uint) error
	Unfollow(followerID, followingID uint) error
	GetStat(userID uint) (*userstatmodel.UserStat, error)
	GetFollowing(userID uint, page, limit int) ([]usermodel.User, int64, error)
	GetFollowers(userID uint, page, limit int) ([]usermodel.User, int64, error)
}
