package repository

import (
	"time"

	"github.com/sunibgram/domain/model"
)

type IRepostRepository interface {
	Save(repost *model.Repost) error
	Delete(userID, threadID uint) error
	Exists(userID, threadID uint) (bool, error)
	AreThreadsReposted(userID uint, threadIDs []uint) (map[uint]bool, error)
	FindByUserID(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error)
}
