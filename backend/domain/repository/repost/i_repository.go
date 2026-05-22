package repostrepoiface

import (
	"time"

	repostmodel "github.com/sunibgram/domain/model/repost"
	threadmodel "github.com/sunibgram/domain/model/thread"
)

type IRepostRepository interface {
	Save(repost *repostmodel.Repost) error
	Delete(userID, threadID uint) error
	Exists(userID, threadID uint) (bool, error)
	AreThreadsReposted(userID uint, threadIDs []uint) (map[uint]bool, error)
	FindByUserID(userID uint, page, limit int) ([]threadmodel.Thread, []time.Time, int64, error)
}
