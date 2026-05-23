package bookmarksvciface

import (
	"time"

	threadmodel "github.com/sunibgram/domain/model/thread"
)

type IBookmarkService interface {
	Bookmark(userID, threadID uint) error
	Unbookmark(userID, threadID uint) error
	IsBookmarked(userID, threadID uint) bool
	AreThreadsBookmarked(userID uint, threadIDs []uint) (map[uint]bool, error)
	AllBookmarkedThreadIDs(userID uint) ([]uint, error)
	GetMyBookmarks(userID uint, page, limit int) ([]threadmodel.Thread, []time.Time, int64, error)
}
