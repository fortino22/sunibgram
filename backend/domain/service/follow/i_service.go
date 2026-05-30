package followsvciface

import (
	usermodel "github.com/sunibgram/domain/model/user"
	userstatmodel "github.com/sunibgram/domain/model/user_stat"
)

type IFollowService interface {
	Follow(followerID, followingID uint) error
	Unfollow(followerID, followingID uint) error
	GetStat(userID uint) (*userstatmodel.UserStat, error)
	GetFollowing(viewerID, userID uint, page, limit int) ([]usermodel.User, int64, error)
	GetFollowers(viewerID, userID uint, page, limit int) ([]usermodel.User, int64, error)
}
