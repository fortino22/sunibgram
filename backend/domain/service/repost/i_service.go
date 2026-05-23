package repostsvciface

import (
	"time"

	threadmodel "github.com/sunibgram/domain/model/thread"
)

type IRepostService interface {
	Repost(userID, threadID uint) error
	Unrepost(userID, threadID uint) error
	AreThreadsReposted(userID uint, threadIDs []uint) (map[uint]bool, error)
	GetMyReposts(userID uint, page, limit int) ([]threadmodel.Thread, []time.Time, int64, error)
}
