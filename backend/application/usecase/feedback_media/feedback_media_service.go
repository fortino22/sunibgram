package feedbackmediauc

import (
	"context"
	"log"
	"mime/multipart"
	"sync"

	"github.com/google/uuid"

	storageuc "github.com/sunibgram/application/usecase/storage"
	feedbackmediamodel "github.com/sunibgram/domain/model/feedback_media"
	mediamodel "github.com/sunibgram/domain/model/media"
	feedbackmediarepoiface "github.com/sunibgram/domain/repository/feedback_media"
	feedbackmediasvciface "github.com/sunibgram/domain/service/feedback_media"
)

type FeedbackMediaService struct {
	mediaRepo feedbackmediarepoiface.IFeedbackMediaRepository
	storage   *storageuc.StorageUsecase
}

func NewFeedbackMediaService(
	mediaRepo feedbackmediarepoiface.IFeedbackMediaRepository,
	storage *storageuc.StorageUsecase,
) feedbackmediasvciface.IFeedbackMediaService {
	return &FeedbackMediaService{mediaRepo: mediaRepo, storage: storage}
}

func (s *FeedbackMediaService) Upload(ctx context.Context, feedbackID uuid.UUID, files []*multipart.FileHeader) ([]feedbackmediamodel.FeedbackMedia, error) {
	if err := feedbackmediamodel.ValidateFeedbackMediaCount(len(files)); err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, nil
	}

	_, prefix, err := s.storage.ResolveBucketAndPrefix("post")
	if err != nil {
		return nil, err
	}

	media, err := s.uploadFiles(ctx, feedbackID, prefix, files)
	if err != nil {
		return nil, err
	}

	if err := s.mediaRepo.SaveAll(media); err != nil {
		return nil, err
	}
	return media, nil
}

func (s *FeedbackMediaService) uploadFiles(ctx context.Context, feedbackID uuid.UUID, prefix string, files []*multipart.FileHeader) ([]feedbackmediamodel.FeedbackMedia, error) {
	for _, header := range files {
		if err := mediamodel.ValidateUploadSize(header, mediamodel.MaxPostUploadSize); err != nil {
			return nil, err
		}
		if err := mediamodel.ValidateImageExt(header.Filename); err != nil {
			return nil, err
		}
	}

	media := make([]feedbackmediamodel.FeedbackMedia, len(files))
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

			if err := mediamodel.ValidateImageFile(file); err != nil {
				errCh <- err
				return
			}

			objectName, err := s.storage.Upload(ctx, prefix, file, header)
			if err != nil {
				errCh <- err
				return
			}

			media[i] = feedbackmediamodel.FeedbackMedia{
				FeedbackID: feedbackID,
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

func (s *FeedbackMediaService) GetByFeedbackID(feedbackID uuid.UUID) ([]feedbackmediamodel.FeedbackMedia, error) {
	return s.mediaRepo.FindByFeedbackID(feedbackID)
}

func (s *FeedbackMediaService) GetByFeedbackIDs(feedbackIDs []uuid.UUID) ([]feedbackmediamodel.FeedbackMedia, error) {
	return s.mediaRepo.FindByFeedbackIDs(feedbackIDs)
}

func (s *FeedbackMediaService) DeleteByFeedback(ctx context.Context, feedbackID uuid.UUID) error {
	media, err := s.mediaRepo.FindByFeedbackID(feedbackID)
	if err != nil {
		return err
	}
	for _, m := range media {
		if err := s.storage.Delete(ctx, m.ObjectName); err != nil {
			log.Printf("feedback media orphan: %s: %v", m.ObjectName, err)
		}
	}
	return s.mediaRepo.DeleteByFeedbackID(feedbackID)
}

func (s *FeedbackMediaService) GetURL(objectName string) string {
	return s.storage.GetURL(objectName)
}
