package usecase

import (
	"context"
	"errors"
	"mime/multipart"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type ThreadMediaService struct {
	threadRepo repository.IThreadRepository
	mediaRepo  repository.IThreadMediaRepository
	storage    *StorageUsecase
}

func NewThreadMediaService(
	threadRepo repository.IThreadRepository,
	mediaRepo repository.IThreadMediaRepository,
	storage *StorageUsecase,
) service.IThreadMediaService {
	return &ThreadMediaService{threadRepo: threadRepo, mediaRepo: mediaRepo, storage: storage}
}

func (s *ThreadMediaService) Upload(ctx context.Context, threadID, userID uint, files []*multipart.FileHeader) ([]model.ThreadMedia, error) {
	thread, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return nil, errors.New("thread not found")
	}
	if thread.UserID != userID {
		return nil, errors.New("forbidden")
	}

	bucket, err := s.storage.ResolveBucket("post")
	if err != nil {
		return nil, err
	}

	var media []model.ThreadMedia
	for _, header := range files {
		if err := model.ValidateThreadMediaExt(header.Filename); err != nil {
			return nil, err
		}

		file, err := header.Open()
		if err != nil {
			return nil, err
		}

		objectName, uploadErr := s.storage.Upload(ctx, bucket, file, header)
		file.Close()
		if uploadErr != nil {
			return nil, uploadErr
		}

		media = append(media, model.ThreadMedia{
			ThreadID:   threadID,
			ObjectName: objectName,
			Bucket:     bucket,
		})
	}

	if err := s.mediaRepo.SaveAll(media); err != nil {
		return nil, err
	}
	return media, nil
}

// Update deletes all existing media for the thread and uploads the new set.
func (s *ThreadMediaService) Update(ctx context.Context, threadID, userID uint, files []*multipart.FileHeader) error {
	thread, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return errors.New("thread not found")
	}
	if thread.UserID != userID {
		return errors.New("forbidden")
	}

	existing, _ := s.mediaRepo.FindByThreadID(threadID)
	for _, m := range existing {
		_ = s.storage.Delete(ctx, m.Bucket, m.ObjectName)
	}
	_ = s.mediaRepo.DeleteByThreadID(threadID)

	bucket, err := s.storage.ResolveBucket("post")
	if err != nil {
		return err
	}

	var media []model.ThreadMedia
	for _, header := range files {
		if err := model.ValidateThreadMediaExt(header.Filename); err != nil {
			return err
		}

		file, err := header.Open()
		if err != nil {
			return err
		}

		objectName, uploadErr := s.storage.Upload(ctx, bucket, file, header)
		file.Close()
		if uploadErr != nil {
			return uploadErr
		}

		media = append(media, model.ThreadMedia{
			ThreadID:   threadID,
			ObjectName: objectName,
			Bucket:     bucket,
		})
	}

	return s.mediaRepo.SaveAll(media)
}

func (s *ThreadMediaService) DeleteByThread(ctx context.Context, threadID uint) error {
	media, err := s.mediaRepo.FindByThreadID(threadID)
	if err != nil {
		return err
	}
	for _, m := range media {
		_ = s.storage.Delete(ctx, m.Bucket, m.ObjectName)
	}
	return s.mediaRepo.DeleteByThreadID(threadID)
}

func (s *ThreadMediaService) GetByThreadID(threadID uint) ([]model.ThreadMedia, error) {
	return s.mediaRepo.FindByThreadID(threadID)
}

func (s *ThreadMediaService) GetByThreadIDs(threadIDs []uint) ([]model.ThreadMedia, error) {
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
		_ = s.storage.Delete(ctx, m.Bucket, m.ObjectName)
	}
	return nil
}

func (s *ThreadMediaService) GetURL(bucket, objectName string) string {
	return s.storage.GetURL(bucket, objectName)
}
