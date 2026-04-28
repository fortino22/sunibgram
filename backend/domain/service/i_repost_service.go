package service

import (
	"time"

	"github.com/sunibgram/domain/model"
)

type IRepostService interface {
	Repost(userID, threadID uint) error
	Unrepost(userID, threadID uint) error
	GetMyReposts(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error)
}
