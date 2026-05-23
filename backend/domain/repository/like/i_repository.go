package likerepoiface

import (
	"time"

	likemodel "github.com/sunibgram/domain/model/like"
	threadmodel "github.com/sunibgram/domain/model/thread"
)

type ILikeRepository interface {
	Save(like *likemodel.Like) error
	Delete(userID, threadID uint) error
	Exists(userID, threadID uint) (bool, error)
	AreThreadsLiked(userID uint, threadIDs []uint) (map[uint]bool, error)
	AllLikedThreadIDs(userID uint) ([]uint, error)
	FindByUserID(userID uint, page, limit int) ([]threadmodel.Thread, []time.Time, int64, error)
}
