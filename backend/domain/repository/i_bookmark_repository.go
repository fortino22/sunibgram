package repository

import (
	"time"

	bookmarkmodel "github.com/sunibgram/domain/model/bookmark"
	threadmodel "github.com/sunibgram/domain/model/thread"
)

type IBookmarkRepository interface {
	Save(bookmark *bookmarkmodel.Bookmark) error
	Delete(userID, threadID uint) error
	Exists(userID, threadID uint) (bool, error)
	AreThreadsBookmarked(userID uint, threadIDs []uint) (map[uint]bool, error)
	AllBookmarkedThreadIDs(userID uint) ([]uint, error)
	FindByUserID(userID uint, page, limit int) ([]threadmodel.Thread, []time.Time, int64, error)
}
