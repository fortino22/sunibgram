package usecase

import (
	"context"
	"errors"
	"mime/multipart"
	"sync"

	chatmediamodel "github.com/sunibgram/domain/model/chat_media"
	mediamodel "github.com/sunibgram/domain/model/media"
	chatrepoiface "github.com/sunibgram/domain/repository/chat"
	chatmediarepoiface "github.com/sunibgram/domain/repository/chat_media"
	chatmediasvciface "github.com/sunibgram/domain/service/chat_media"
)

type ChatMediaService struct {
	chatRepo  chatrepoiface.IChatRepository
	mediaRepo chatmediarepoiface.IChatMediaRepository
	storage   *StorageUsecase
}

func NewChatMediaService(chatRepo chatrepoiface.IChatRepository, mediaRepo chatmediarepoiface.IChatMediaRepository, storage *StorageUsecase) chatmediasvciface.IChatMediaService {
	return &ChatMediaService{chatRepo: chatRepo, mediaRepo: mediaRepo, storage: storage}
}

func (s *ChatMediaService) Upload(ctx context.Context, chatID, senderID uint, files []*multipart.FileHeader) ([]chatmediamodel.ChatMedia, error) {
	if err := chatmediamodel.ValidateChatMediaCount(len(files)); err != nil {
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
		if err := mediamodel.ValidateUploadSize(header, mediamodel.MaxChatMediaUploadSize); err != nil {
			return nil, err
		}
		if err := chatmediamodel.ValidateChatMediaExt(header.Filename); err != nil {
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

func (s *ChatMediaService) uploadFiles(ctx context.Context, chatID uint, bucket string, files []*multipart.FileHeader) ([]chatmediamodel.ChatMedia, error) {
	media := make([]chatmediamodel.ChatMedia, len(files))
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

			mediaType, err := chatmediamodel.ValidateChatMediaFile(file)
			if err != nil {
				errCh <- err
				return
			}

			objectName, err := s.storage.Upload(ctx, bucket, file, header)
			if err != nil {
				errCh <- err
				return
			}

			media[i] = chatmediamodel.ChatMedia{
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

func (s *ChatMediaService) rollbackUploaded(ctx context.Context, chatID uint, media []chatmediamodel.ChatMedia) {
	_ = s.mediaRepo.DeleteByChatID(chatID)
	for _, item := range media {
		if item.ObjectName == "" {
			continue
		}
		_ = s.storage.Delete(ctx, item.Bucket, item.ObjectName)
	}
}

func (s *ChatMediaService) GetByChatID(chatID uint) ([]chatmediamodel.ChatMedia, error) {
	return s.mediaRepo.FindByChatID(chatID)
}

func (s *ChatMediaService) GetByChatIDs(chatIDs []uint) ([]chatmediamodel.ChatMedia, error) {
	return s.mediaRepo.FindByChatIDs(chatIDs)
}

func (s *ChatMediaService) GetURL(bucket, objectName string) string {
	return s.storage.GetURL(bucket, objectName)
}

func (s *ChatMediaService) GeneratePresignedURL(ctx context.Context, bucket, objectName string) (string, error) {
	return s.storage.GeneratePresignedURL(ctx, bucket, objectName)
}
