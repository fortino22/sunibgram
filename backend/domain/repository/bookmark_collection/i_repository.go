package bookmarkcollectionrepoiface

import (
	"time"

	bookmarkcollectionmodel "github.com/sunibgram/domain/model/bookmark_collection"
	threadmodel "github.com/sunibgram/domain/model/thread"
)

type IBookmarkCollectionRepository interface {
	Create(collection *bookmarkcollectionmodel.BookmarkCollection) error
	Update(collectionID uint, name string, description *string) error
	Delete(collectionID uint) error
	FindByID(collectionID uint) (*bookmarkcollectionmodel.BookmarkCollection, error)
	FindByUserID(userID uint) ([]bookmarkcollectionmodel.BookmarkCollection, error)
	CountByUserID(userID uint) (int64, error)
	ExistsByUserAndName(userID uint, name string, excludeID uint) (bool, error)

	AddItem(item *bookmarkcollectionmodel.BookmarkCollectionItem) error
	RemoveItem(collectionID, threadID uint) error
	RemoveItemsByBookmark(userID, threadID uint) error
	ItemExists(collectionID, threadID uint) (bool, error)
	CollectionIDsForBookmark(userID, threadID uint) ([]uint, error)
	FindThreadsByCollection(collectionID uint, page, limit int) ([]threadmodel.Thread, []time.Time, int64, error)
}
