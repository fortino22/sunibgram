package bookmarkcollectionrepo

import (
	"time"

	"gorm.io/gorm"

	bookmarkcollectionmodel "github.com/sunibgram/domain/model/bookmark_collection"
	threadmodel "github.com/sunibgram/domain/model/thread"
	bookmarkcollectionrepoiface "github.com/sunibgram/domain/repository/bookmark_collection"
	threadrepo "github.com/sunibgram/infrastructure/repository/thread"
)

type bookmarkCollectionRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewBookmarkCollectionRepository(writeDB, readDB *gorm.DB) bookmarkcollectionrepoiface.IBookmarkCollectionRepository {
	return &bookmarkCollectionRepository{writeDB: writeDB, readDB: readDB}
}

func (r *bookmarkCollectionRepository) Create(c *bookmarkcollectionmodel.BookmarkCollection) error {
	return r.writeDB.Create(c).Error
}

func (r *bookmarkCollectionRepository) Update(collectionID uint, name string, description *string) error {
	updates := map[string]any{
		"name":        name,
		"description": description,
		"updated_at":  time.Now(),
	}
	return r.writeDB.Model(&bookmarkcollectionmodel.BookmarkCollection{}).
		Where("id = ?", collectionID).Updates(updates).Error
}

func (r *bookmarkCollectionRepository) Delete(collectionID uint) error {
	return r.writeDB.Where("id = ?", collectionID).Delete(&bookmarkcollectionmodel.BookmarkCollection{}).Error
}

func (r *bookmarkCollectionRepository) FindByID(collectionID uint) (*bookmarkcollectionmodel.BookmarkCollection, error) {
	var c bookmarkcollectionmodel.BookmarkCollection
	if err := r.readDB.Where("id = ?", collectionID).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *bookmarkCollectionRepository) FindByUserID(userID uint) ([]bookmarkcollectionmodel.BookmarkCollection, error) {
	var list []bookmarkcollectionmodel.BookmarkCollection
	// Defensive cap: a user's collection count is small by UX, but bound the
	// query so a pathological account can't load an unbounded row set.
	if err := r.readDB.Where("user_id = ?", userID).Order("created_at DESC").Limit(200).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *bookmarkCollectionRepository) CountByUserID(userID uint) (int64, error) {
	var n int64
	err := r.readDB.Model(&bookmarkcollectionmodel.BookmarkCollection{}).
		Where("user_id = ?", userID).Count(&n).Error
	return n, err
}

func (r *bookmarkCollectionRepository) ExistsByUserAndName(userID uint, name string, excludeID uint) (bool, error) {
	var exists bool
	q := r.readDB.Model(&bookmarkcollectionmodel.BookmarkCollection{}).
		Select("1").
		Where("user_id = ? AND name = ?", userID, name)
	if excludeID != 0 {
		q = q.Where("id <> ?", excludeID)
	}
	err := q.Limit(1).Scan(&exists).Error
	return exists, err
}

func (r *bookmarkCollectionRepository) AddItem(item *bookmarkcollectionmodel.BookmarkCollectionItem) error {
	return r.writeDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(item).Error; err != nil {
			return err
		}
		return tx.Model(&bookmarkcollectionmodel.BookmarkCollection{}).
			Where("id = ?", item.CollectionID).
			UpdateColumn("item_count", gorm.Expr("item_count + 1")).Error
	})
}

func (r *bookmarkCollectionRepository) RemoveItem(collectionID, threadID uint) error {
	return r.writeDB.Transaction(func(tx *gorm.DB) error {
		res := tx.Where("collection_id = ? AND thread_id = ?", collectionID, threadID).
			Delete(&bookmarkcollectionmodel.BookmarkCollectionItem{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil
		}
		return tx.Model(&bookmarkcollectionmodel.BookmarkCollection{}).
			Where("id = ?", collectionID).
			UpdateColumn("item_count", gorm.Expr("GREATEST(item_count - 1, 0)")).Error
	})
}

func (r *bookmarkCollectionRepository) RemoveItemsByBookmark(userID, threadID uint) error {
	return r.writeDB.Transaction(func(tx *gorm.DB) error {
		var affectedIDs []uint
		if err := tx.Model(&bookmarkcollectionmodel.BookmarkCollectionItem{}).
			Where("user_id = ? AND thread_id = ?", userID, threadID).
			Pluck("collection_id", &affectedIDs).Error; err != nil {
			return err
		}
		if len(affectedIDs) == 0 {
			return nil
		}
		if err := tx.Where("user_id = ? AND thread_id = ?", userID, threadID).
			Delete(&bookmarkcollectionmodel.BookmarkCollectionItem{}).Error; err != nil {
			return err
		}
		return tx.Model(&bookmarkcollectionmodel.BookmarkCollection{}).
			Where("id IN ?", affectedIDs).
			UpdateColumn("item_count", gorm.Expr("GREATEST(item_count - 1, 0)")).Error
	})
}

func (r *bookmarkCollectionRepository) ItemExists(collectionID, threadID uint) (bool, error) {
	var exists bool
	err := r.readDB.Model(&bookmarkcollectionmodel.BookmarkCollectionItem{}).
		Select("1").
		Where("collection_id = ? AND thread_id = ?", collectionID, threadID).
		Limit(1).Scan(&exists).Error
	return exists, err
}

func (r *bookmarkCollectionRepository) FindThreadsByCollection(collectionID uint, page, limit int) ([]threadmodel.Thread, []time.Time, int64, error) {
	type row struct {
		ThreadID  uint
		CreatedAt time.Time
	}

	var total int64
	if err := r.readDB.Model(&bookmarkcollectionmodel.BookmarkCollectionItem{}).
		Where("collection_id = ?", collectionID).Count(&total).Error; err != nil {
		return nil, nil, 0, err
	}

	var rows []row
	offset := (page - 1) * limit
	if err := r.readDB.Model(&bookmarkcollectionmodel.BookmarkCollectionItem{}).
		Select("thread_id, created_at").
		Where("collection_id = ?", collectionID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, nil, total, err
	}
	if len(rows) == 0 {
		return nil, nil, total, nil
	}

	ids := make([]uint, len(rows))
	timeMap := make(map[uint]time.Time, len(rows))
	for i, x := range rows {
		ids[i] = x.ThreadID
		timeMap[x.ThreadID] = x.CreatedAt
	}

	var threads []threadmodel.Thread
	if err := threadrepo.PreloadThreadDetails(r.readDB).
		Where("threads.thread_id IN ?", ids).Find(&threads).Error; err != nil {
		return nil, nil, total, err
	}

	tmap := make(map[uint]threadmodel.Thread, len(threads))
	for _, t := range threads {
		tmap[t.ThreadID] = t
	}
	sorted := make([]threadmodel.Thread, 0, len(ids))
	stamps := make([]time.Time, 0, len(ids))
	for _, id := range ids {
		if t, ok := tmap[id]; ok {
			sorted = append(sorted, t)
			stamps = append(stamps, timeMap[id])
		}
	}
	return sorted, stamps, total, nil
}
