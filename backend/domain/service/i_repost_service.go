package service

import (
	"time"

	"github.com/sunibgram/domain/model"
)

type IRepostService interface {
	Repost(userID, threadID uint) error
	Unrepost(userID, threadID uint) error
	AreThreadsReposted(userID uint, threadIDs []uint) (map[uint]bool, error)
	GetMyReposts(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error)
}
