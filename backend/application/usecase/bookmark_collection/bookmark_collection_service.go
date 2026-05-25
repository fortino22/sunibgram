package bookmarkcollectionuc

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/sunibgram/application/helper"
	threaduc "github.com/sunibgram/application/usecase/thread"
	bookmarkcollectionmodel "github.com/sunibgram/domain/model/bookmark_collection"
	threadmodel "github.com/sunibgram/domain/model/thread"
	bookmarkrepoiface "github.com/sunibgram/domain/repository/bookmark"
	bookmarkcollectionrepoiface "github.com/sunibgram/domain/repository/bookmark_collection"
	followrepoiface "github.com/sunibgram/domain/repository/follow"
	threadrepoiface "github.com/sunibgram/domain/repository/thread"
	usersettingsrepoiface "github.com/sunibgram/domain/repository/user_settings"
	bookmarkcollectionsvciface "github.com/sunibgram/domain/service/bookmark_collection"
)

type BookmarkCollectionService struct {
	collectionRepo bookmarkcollectionrepoiface.IBookmarkCollectionRepository
	bookmarkRepo   bookmarkrepoiface.IBookmarkRepository
	threadRepo     threadrepoiface.IThreadRepository
	settingsRepo   usersettingsrepoiface.IUserSettingsRepository
	followRepo     followrepoiface.IFollowRepository
}

func NewBookmarkCollectionService(
	collectionRepo bookmarkcollectionrepoiface.IBookmarkCollectionRepository,
	bookmarkRepo bookmarkrepoiface.IBookmarkRepository,
	threadRepo threadrepoiface.IThreadRepository,
	settingsRepo usersettingsrepoiface.IUserSettingsRepository,
	followRepo followrepoiface.IFollowRepository,
) bookmarkcollectionsvciface.IBookmarkCollectionService {
	return &BookmarkCollectionService{
		collectionRepo: collectionRepo,
		bookmarkRepo:   bookmarkRepo,
		threadRepo:     threadRepo,
		settingsRepo:   settingsRepo,
		followRepo:     followRepo,
	}
}

func validateName(name string) (string, error) {
	n := strings.TrimSpace(name)
	if n == "" {
		return "", errors.New("name is required")
	}
	if len(n) > bookmarkcollectionmodel.MaxCollectionNameLen {
		return "", errors.New("name too long")
	}
	return n, nil
}

func (s *BookmarkCollectionService) Create(userID uint, name string, description *string) (*bookmarkcollectionmodel.BookmarkCollection, error) {
	n, err := validateName(name)
	if err != nil {
		return nil, err
	}
	count, err := s.collectionRepo.CountByUserID(userID)
	if err != nil {
		return nil, err
	}
	if count >= int64(bookmarkcollectionmodel.MaxCollectionsPerUser) {
		return nil, errors.New("collection limit reached")
	}
	dup, err := s.collectionRepo.ExistsByUserAndName(userID, n, 0)
	if err != nil {
		return nil, err
	}
	if dup {
		return nil, errors.New("collection name already exists")
	}
	c := &bookmarkcollectionmodel.BookmarkCollection{
		UserID:      userID,
		Name:        n,
		Description: description,
	}
	if err := s.collectionRepo.Create(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *BookmarkCollectionService) ownCollection(userID, collectionID uint) (*bookmarkcollectionmodel.BookmarkCollection, error) {
	c, err := s.collectionRepo.FindByID(collectionID)
	if err != nil {
		return nil, errors.New("collection not found")
	}
	if c.UserID != userID {
		return nil, errors.New("forbidden")
	}
	return c, nil
}

func (s *BookmarkCollectionService) Update(userID, collectionID uint, name string, description *string) error {
	if _, err := s.ownCollection(userID, collectionID); err != nil {
		return err
	}
	n, err := validateName(name)
	if err != nil {
		return err
	}
	dup, err := s.collectionRepo.ExistsByUserAndName(userID, n, collectionID)
	if err != nil {
		return err
	}
	if dup {
		return errors.New("collection name already exists")
	}
	return s.collectionRepo.Update(collectionID, n, description)
}

func (s *BookmarkCollectionService) Delete(userID, collectionID uint) error {
	if _, err := s.ownCollection(userID, collectionID); err != nil {
		return err
	}
	return s.collectionRepo.Delete(collectionID)
}

func (s *BookmarkCollectionService) ListMine(userID uint) ([]bookmarkcollectionmodel.BookmarkCollection, error) {
	return s.collectionRepo.FindByUserID(userID)
}

func (s *BookmarkCollectionService) GetMine(userID, collectionID uint) (*bookmarkcollectionmodel.BookmarkCollection, error) {
	return s.ownCollection(userID, collectionID)
}

func (s *BookmarkCollectionService) AddThread(userID, collectionID, threadID uint) error {
	if _, err := s.ownCollection(userID, collectionID); err != nil {
		return err
	}
	thread, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return errors.New("thread not found")
	}
	if err := threaduc.CanViewPrivateUser(context.Background(), s.settingsRepo, s.followRepo, &userID, thread.UserID); err != nil {
		return err
	}
	bookmarked, err := s.bookmarkRepo.Exists(userID, threadID)
	if err != nil {
		return err
	}
	if !bookmarked {
		return errors.New("thread not bookmarked")
	}
	exists, err := s.collectionRepo.ItemExists(collectionID, threadID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("thread already in collection")
	}
	return s.collectionRepo.AddItem(&bookmarkcollectionmodel.BookmarkCollectionItem{
		CollectionID: collectionID,
		UserID:       userID,
		ThreadID:     threadID,
		CreatedAt:    time.Now(),
	})
}

func (s *BookmarkCollectionService) RemoveThread(userID, collectionID, threadID uint) error {
	if _, err := s.ownCollection(userID, collectionID); err != nil {
		return err
	}
	exists, err := s.collectionRepo.ItemExists(collectionID, threadID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("thread not in collection")
	}
	return s.collectionRepo.RemoveItem(collectionID, threadID)
}

func (s *BookmarkCollectionService) CollectionIDsForThread(userID, threadID uint) ([]uint, error) {
	return s.collectionRepo.CollectionIDsForBookmark(userID, threadID)
}

func (s *BookmarkCollectionService) ListThreads(userID, collectionID uint, page, limit int) ([]threadmodel.Thread, []time.Time, int64, error) {
	if _, err := s.ownCollection(userID, collectionID); err != nil {
		return nil, nil, 0, err
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.collectionRepo.FindThreadsByCollection(collectionID, page, limit)
}
