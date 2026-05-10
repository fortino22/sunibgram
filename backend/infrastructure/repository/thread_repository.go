package repository

import (
	"time"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
)

type threadCached struct {
	db        *threadDB
	cache     *cache.Cache
	ttl       time.Duration
	likes     *interactionRepo
	bookmarks *interactionRepo
	reposts   *interactionRepo
}

type threadDB struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

const threadFeedTTL = 60 * time.Second

func preloadThreadUser(db *gorm.DB) *gorm.DB {
	return db.Select("user_id", "name", "user_name", "is_premium")
}

func (db *threadDB) attachFaculties(threads []model.Thread) {
	userIDs := make(map[uint]struct{})
	for i := range threads {
		userIDs[threads[i].UserID] = struct{}{}
		if threads[i].OriginalThread != nil {
			userIDs[threads[i].OriginalThread.UserID] = struct{}{}
		}
	}
	if len(userIDs) == 0 {
		return
	}
	ids := make([]uint, 0, len(userIDs))
	for id := range userIDs {
		ids = append(ids, id)
	}
	var faculties []model.UserFaculty
	db.readDB.Preload("Faculty").Where("user_id IN ?", ids).Find(&faculties)
	fm := make(map[uint]*model.UserFaculty, len(faculties))
	for i := range faculties {
		fm[faculties[i].UserID] = &faculties[i]
	}
	for i := range threads {
		if uf, ok := fm[threads[i].UserID]; ok {
			threads[i].User.UserFaculty = uf
		}
		if threads[i].OriginalThread != nil {
			if uf, ok := fm[threads[i].OriginalThread.UserID]; ok {
				threads[i].OriginalThread.User.UserFaculty = uf
			}
		}
	}
}

func NewThreadRepository(writeDB, readDB *gorm.DB, c *cache.Cache) (domainrepo.IThreadRepository, func(uint, string, int)) {
	tdb := &threadDB{writeDB: writeDB, readDB: readDB}
	r := &threadCached{
		db:        tdb,
		cache:     c,
		ttl:       cache.DefaultTTL,
		likes:     newInteractionRepo(writeDB, readDB, c, "likes", tdb),
		bookmarks: newInteractionRepo(writeDB, readDB, c, "bookmarks", tdb),
		reposts:   newInteractionRepo(writeDB, readDB, c, "reposts", tdb),
	}
	return r, r.patchCachedCount
}
