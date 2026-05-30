package threadmediauc

import (
	"context"
	"errors"
	"log"
	"mime/multipart"
	"sync"

	storageuc "github.com/sunibgram/application/usecase/storage"
	mediamodel "github.com/sunibgram/domain/model/media"
	threadmediamodel "github.com/sunibgram/domain/model/thread_media"
	threadrepoiface "github.com/sunibgram/domain/repository/thread"
	threadmediarepoiface "github.com/sunibgram/domain/repository/thread_media"
	threadmediasvciface "github.com/sunibgram/domain/service/thread_media"
)

type ThreadMediaService struct {
	threadRepo threadrepoiface.IThreadRepository
	mediaRepo  threadmediarepoiface.IThreadMediaRepository
	storage    *storageuc.StorageUsecase
}

func NewThreadMediaService(
	threadRepo threadrepoiface.IThreadRepository,
	mediaRepo threadmediarepoiface.IThreadMediaRepository,
	storage *storageuc.StorageUsecase,
) threadmediasvciface.IThreadMediaService {
	return &ThreadMediaService{threadRepo: threadRepo, mediaRepo: mediaRepo, storage: storage}
}

func (s *ThreadMediaService) Upload(ctx context.Context, threadID, userID uint, files []*multipart.FileHeader) ([]threadmediamodel.ThreadMedia, error) {
	if err := threadmediamodel.ValidateThreadMediaCount(len(files)); err != nil {
		return nil, err
	}

	thread, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return nil, errors.New("thread not found")
	}
	if thread.UserID != userID {
		return nil, errors.New("forbidden")
	}

	_, prefix, err := s.storage.ResolveBucketAndPrefix("post")
	if err != nil {
		return nil, err
	}

	media, err := s.uploadFiles(ctx, threadID, prefix, files)
	if err != nil {
		return nil, err
	}

	if err := s.mediaRepo.SaveAll(media); err != nil {
		return nil, err
	}
	return media, nil
}

// Update deletes all existing media for the thread and uploads the new set.
func (s *ThreadMediaService) Update(ctx context.Context, threadID, userID uint, files []*multipart.FileHeader) error {
	if err := threadmediamodel.ValidateThreadMediaCount(len(files)); err != nil {
		return err
	}

	thread, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return errors.New("thread not found")
	}
	if thread.UserID != userID {
		return errors.New("forbidden")
	}

	existing, _ := s.mediaRepo.FindByThreadID(threadID)
	for _, m := range existing {
		if err := s.storage.Delete(ctx, m.ObjectName); err != nil {
			log.Printf("thread media orphan: %s: %v", m.ObjectName, err)
		}
	}
	_ = s.mediaRepo.DeleteByThreadID(threadID)

	_, prefix, err := s.storage.ResolveBucketAndPrefix("post")
	if err != nil {
		return err
	}

	media, err := s.uploadFiles(ctx, threadID, prefix, files)
	if err != nil {
		return err
	}

	return s.mediaRepo.SaveAll(media)
}

func (s *ThreadMediaService) UpdateGranular(ctx context.Context, threadID, userID uint, keepIDs []uint, files []*multipart.FileHeader) error {
	if err := threadmediamodel.ValidateThreadMediaCount(len(keepIDs) + len(files)); err != nil {
		return err
	}

	thread, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return errors.New("thread not found")
	}
	if thread.UserID != userID {
		return errors.New("forbidden")
	}

	existing, err := s.mediaRepo.FindByThreadID(threadID)
	if err != nil {
		return err
	}

	keepSet := make(map[uint]struct{}, len(keepIDs))
	for _, id := range keepIDs {
		keepSet[id] = struct{}{}
	}

	existingSet := make(map[uint]threadmediamodel.ThreadMedia, len(existing))
	for _, m := range existing {
		existingSet[m.MediaID] = m
	}
	for _, id := range keepIDs {
		if _, ok := existingSet[id]; !ok {
			return errors.New("media not found")
		}
	}

	deleteIDs := make([]uint, 0, len(existing))
	for _, m := range existing {
		if _, keep := keepSet[m.MediaID]; !keep {
			deleteIDs = append(deleteIDs, m.MediaID)
		}
	}

	if len(deleteIDs) > 0 {
		deleted, err := s.mediaRepo.DeleteByIDs(deleteIDs)
		if err != nil {
			return err
		}
		for _, m := range deleted {
			if err := s.storage.Delete(ctx, m.ObjectName); err != nil {
			log.Printf("thread media orphan: %s: %v", m.ObjectName, err)
		}
		}
	}

	if len(files) == 0 {
		return nil
	}

	_, prefix, err := s.storage.ResolveBucketAndPrefix("post")
	if err != nil {
		return err
	}

	media, err := s.uploadFiles(ctx, threadID, prefix, files)
	if err != nil {
		return err
	}

	return s.mediaRepo.SaveAll(media)
}

func (s *ThreadMediaService) uploadFiles(ctx context.Context, threadID uint, prefix string, files []*multipart.FileHeader) ([]threadmediamodel.ThreadMedia, error) {
	for _, header := range files {
		if err := mediamodel.ValidateUploadSize(header, mediamodel.MaxPostUploadSize); err != nil {
			return nil, err
		}
		if err := threadmediamodel.ValidateThreadMediaExt(header.Filename); err != nil {
			return nil, err
		}
	}

	media := make([]threadmediamodel.ThreadMedia, len(files))
	errCh := make(chan error, len(files))
	var wg sync.WaitGroup

	for i, header := range files {
		i, header := i, header
		wg.Add(1)
		go func() {
			defer wg.Done()

			file, err := header.Open()
			if err != nil {
				errCh <- err
				return
			}
			defer file.Close()

			if _, err := threadmediamodel.ValidateThreadMediaFile(file); err != nil {
				errCh <- err
				return
			}

			objectName, err := s.storage.Upload(ctx, prefix, file, header)
			if err != nil {
				errCh <- err
				return
			}

			media[i] = threadmediamodel.ThreadMedia{
				ThreadID:   threadID,
				ObjectName: objectName,
			}
		}()
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}
	return media, nil
}

func (s *ThreadMediaService) DeleteByThread(ctx context.Context, threadID uint) error {
	media, err := s.mediaRepo.FindByThreadID(threadID)
	if err != nil {
		return err
	}
	for _, m := range media {
		if err := s.storage.Delete(ctx, m.ObjectName); err != nil {
			log.Printf("thread media orphan: %s: %v", m.ObjectName, err)
		}
	}
	return s.mediaRepo.DeleteByThreadID(threadID)
}

func (s *ThreadMediaService) GetByThreadID(threadID uint) ([]threadmediamodel.ThreadMedia, error) {
	return s.mediaRepo.FindByThreadID(threadID)
}

func (s *ThreadMediaService) GetByThreadIDs(threadIDs []uint) ([]threadmediamodel.ThreadMedia, error) {
	return s.mediaRepo.FindByThreadIDs(threadIDs)
}

func (s *ThreadMediaService) Delete(ctx context.Context, threadID, userID uint, mediaIDs []uint) error {
	thread, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return errors.New("thread not found")
	}
	if thread.UserID != userID {
		return errors.New("forbidden")
	}

	// Verify every requested ID exists and belongs to this thread.
	existing, err := s.mediaRepo.FindByThreadID(threadID)
	if err != nil {
		return err
	}
	existingSet := make(map[uint]struct{}, len(existing))
	for _, m := range existing {
		existingSet[m.MediaID] = struct{}{}
	}
	for _, id := range mediaIDs {
		if _, ok := existingSet[id]; !ok {
			return errors.New("media not found")
		}
	}

	deleted, err := s.mediaRepo.DeleteByIDs(mediaIDs)
	if err != nil {
		return err
	}
	for _, m := range deleted {
		if err := s.storage.Delete(ctx, m.ObjectName); err != nil {
			log.Printf("thread media orphan: %s: %v", m.ObjectName, err)
		}
	}
	return nil
}

func (s *ThreadMediaService) GetURL(objectName string) string {
	return s.storage.GetURL(objectName)
}
