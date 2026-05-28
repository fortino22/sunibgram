package followrepoiface

import (
	"context"

	exploremodel "github.com/sunibgram/domain/model/explore"
	usermodel "github.com/sunibgram/domain/model/user"
	userstatmodel "github.com/sunibgram/domain/model/user_stat"
)

type IFollowRepository interface {
	Follow(followerID, followingID uint) error
	Unfollow(followerID, followingID uint) error
	Exists(followerID, followingID uint) (bool, error)
	GetFollowingSet(ctx context.Context, followerID uint, targetIDs []uint) (map[uint]bool, error)
	SearchMentionUsers(ctx context.Context, query string) ([]exploremodel.ExploreUserResult, error)
	SearchMutualFollowIDs(ctx context.Context, viewerID uint, query string, limit int) ([]uint, error)
	HydrateCompareUsers(ctx context.Context, ids []uint) ([]exploremodel.ExploreUserResult, error)
	IsMutualFollow(ctx context.Context, viewerID, targetID uint) (bool, error)
	GetStat(userID uint) (*userstatmodel.UserStat, error)
	GetFollowing(userID uint, page, limit int) ([]usermodel.User, int64, error)
	GetFollowers(userID uint, page, limit int) ([]usermodel.User, int64, error)
}
