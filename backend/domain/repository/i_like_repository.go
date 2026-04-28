package repository

import (
	"time"

	"github.com/sunibgram/domain/model"
)

type ILikeRepository interface {
	Save(like *model.Like) error
	Delete(userID, threadID uint) error
	Exists(userID, threadID uint) (bool, error)
	FindByUserID(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error)
}
