package usecase

import (
	"context"
	"errors"
	"mime/multipart"
	"sync"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type ChatMediaService struct {
	chatRepo  repository.IChatRepository
	mediaRepo repository.IChatMediaRepository
	storage   *StorageUsecase
}

func NewChatMediaService(chatRepo repository.IChatRepository, mediaRepo repository.IChatMediaRepository, storage *StorageUsecase) service.IChatMediaService {
	return &ChatMediaService{chatRepo: chatRepo, mediaRepo: mediaRepo, storage: storage}
}

func (s *ChatMediaService) Upload(ctx context.Context, chatID, senderID uint, files []*multipart.FileHeader) ([]model.ChatMedia, error) {
	if err := model.ValidateChatMediaCount(len(files)); err != nil {
		return nil, err
	}
	chat, err := s.chatRepo.FindByID(chatID)
	if err != nil {
		return nil, errors.New("chat not found")
	}
	if chat.SenderID != senderID {
		return nil, errors.New("forbidden")
	}

	bucket, err := s.storage.ResolveBucket("chat")
	if err != nil {
		return nil, err
	}

	for _, header := range files {
		if err := model.ValidateChatMediaExt(header.Filename); err != nil {
			return nil, err
		}
		if err := model.ValidateChatMediaContentType(header.Header.Get("Content-Type")); err != nil {
			return nil, err
		}
	}

	media, err := s.uploadFiles(ctx, chatID, bucket, files)
	if err != nil {
		s.rollbackUploaded(ctx, chatID, media)
		return nil, err
	}

	for i := range media {
		item := media[i]
		if err := s.mediaRepo.Save(&item); err != nil {
			s.rollbackUploaded(ctx, chatID, media)
			return nil, err
		}
		media[i] = item
	}
	return media, nil
}

func (s *ChatMediaService) uploadFiles(ctx context.Context, chatID uint, bucket string, files []*multipart.FileHeader) ([]model.ChatMedia, error) {
	media := make([]model.ChatMedia, len(files))
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

			objectName, err := s.storage.Upload(ctx, bucket, file, header)
			if err != nil {
				errCh <- err
				return
			}

			mediaType := header.Header.Get("Content-Type")
			if mediaType == "" {
				mediaType = "application/octet-stream"
			}
			media[i] = model.ChatMedia{
				ChatID:     chatID,
				ObjectName: objectName,
				Bucket:     bucket,
				MediaType:  mediaType,
				FileSize:   header.Size,
			}
		}()
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return media, err
		}
	}
	return media, nil
}

func (s *ChatMediaService) rollbackUploaded(ctx context.Context, chatID uint, media []model.ChatMedia) {
	_ = s.mediaRepo.DeleteByChatID(chatID)
	for _, item := range media {
		if item.ObjectName == "" {
			continue
		}
		_ = s.storage.Delete(ctx, item.Bucket, item.ObjectName)
	}
}

func (s *ChatMediaService) GetByChatID(chatID uint) ([]model.ChatMedia, error) {
	return s.mediaRepo.FindByChatID(chatID)
}

func (s *ChatMediaService) GetByChatIDs(chatIDs []uint) ([]model.ChatMedia, error) {
	return s.mediaRepo.FindByChatIDs(chatIDs)
}

func (s *ChatMediaService) GetURL(bucket, objectName string) string {
	return s.storage.GetURL(bucket, objectName)
}

func (s *ChatMediaService) GeneratePresignedURL(ctx context.Context, bucket, objectName string) (string, error) {
	return s.storage.GeneratePresignedURL(ctx, bucket, objectName)
}
