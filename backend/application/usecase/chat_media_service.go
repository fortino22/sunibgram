package usecase

import (
	"context"
	"errors"
	"mime/multipart"

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

	media := make([]model.ChatMedia, 0, len(files))
	for _, header := range files {
		if err := model.ValidateChatMediaExt(header.Filename); err != nil {
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

		item := model.ChatMedia{
			ChatID:     chatID,
			ObjectName: objectName,
			Bucket:     bucket,
			MediaType:  header.Header.Get("Content-Type"),
			FileSize:   header.Size,
		}
		if item.MediaType == "" {
			item.MediaType = "application/octet-stream"
		}
		if err := s.mediaRepo.Save(&item); err != nil {
			return nil, err
		}
		media = append(media, item)
	}
	return media, nil
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
