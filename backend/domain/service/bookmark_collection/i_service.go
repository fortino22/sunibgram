package bookmarkcollectionsvciface

import (
	"time"

	bookmarkcollectionmodel "github.com/sunibgram/domain/model/bookmark_collection"
	threadmodel "github.com/sunibgram/domain/model/thread"
)

type IBookmarkCollectionService interface {
	Create(userID uint, name string, description *string) (*bookmarkcollectionmodel.BookmarkCollection, error)
	Update(userID, collectionID uint, name string, description *string) error
	Delete(userID, collectionID uint) error
	ListMine(userID uint) ([]bookmarkcollectionmodel.BookmarkCollection, error)
	GetMine(userID, collectionID uint) (*bookmarkcollectionmodel.BookmarkCollection, error)

	AddThread(userID, collectionID, threadID uint) error
	RemoveThread(userID, collectionID, threadID uint) error
	ListThreads(userID, collectionID uint, page, limit int) ([]threadmodel.Thread, []time.Time, int64, error)
}
