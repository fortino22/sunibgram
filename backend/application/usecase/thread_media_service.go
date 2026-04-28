package usecase

import (
	"context"
	"errors"
	"mime/multipart"
	"path/filepath"

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

		url, uploadErr := s.storage.Upload(ctx, bucket, file, header)
		file.Close()
		if uploadErr != nil {
			return nil, uploadErr
		}

		media = append(media, model.ThreadMedia{
			ThreadID:   threadID,
			ObjectName: filepath.Base(url),
			Bucket:     bucket,
			URL:        url,
		})
	}

	if err := s.mediaRepo.SaveAll(media); err != nil {
		return nil, err
	}
	return media, nil
}

// Update replaces individual media objects in-place, keeping the same media_id.
// replacements maps each existing media_id to the new file that should replace it.
func (s *ThreadMediaService) Update(ctx context.Context, threadID, userID uint, replacements map[uint]*multipart.FileHeader) error {
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

	existingByID := make(map[uint]model.ThreadMedia, len(existing))
	for _, m := range existing {
		existingByID[m.MediaID] = m
	}

	// Validate all requested IDs exist before touching anything.
	for mediaID := range replacements {
		if _, ok := existingByID[mediaID]; !ok {
			return errors.New("media not found")
		}
	}

	bucket, err := s.storage.ResolveBucket("post")
	if err != nil {
		return err
	}

	for mediaID, header := range replacements {
		if err := model.ValidateThreadMediaExt(header.Filename); err != nil {
			return err
		}

		file, err := header.Open()
		if err != nil {
			return err
		}

		url, uploadErr := s.storage.Upload(ctx, bucket, file, header)
		file.Close()
		if uploadErr != nil {
			return uploadErr
		}

		// Delete old object from MinIO.
		old := existingByID[mediaID]
		_ = s.storage.Delete(ctx, old.Bucket, old.ObjectName)

		// Update the row in-place — same media_id.
		updated := old
		updated.ObjectName = filepath.Base(url)
		updated.Bucket = bucket
		updated.URL = url
		if err := s.mediaRepo.UpdateOne(&updated); err != nil {
			return err
		}
	}

	return nil
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
