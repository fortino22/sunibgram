package service

import (
	"time"

	"github.com/sunibgram/domain/model"
)

type IBookmarkService interface {
	Bookmark(userID, threadID uint) error
	Unbookmark(userID, threadID uint) error
	GetMyBookmarks(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error)
}
