package service

import (
	"time"

	"github.com/sunibgram/domain/model"
)

type ILikeService interface {
	Like(userID, threadID uint) error
	Unlike(userID, threadID uint) error
	IsLiked(userID, threadID uint) bool
	GetMyLikes(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error)
}
