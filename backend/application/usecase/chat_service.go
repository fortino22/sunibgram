package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type ChatService struct {
	chatRepo repository.IChatRepository
	roomRepo repository.IChatRoomRepository
}

func NewChatService(chatRepo repository.IChatRepository, roomRepo repository.IChatRoomRepository) service.IChatService {
	return &ChatService{chatRepo: chatRepo, roomRepo: roomRepo}
}

func (s *ChatService) Send(ctx context.Context, roomID, senderID uint, message *string) (*model.Chat, error) {
	if !s.roomRepo.IsMember(roomID, senderID) {
		return nil, errors.New("forbidden")
	}
	if message != nil {
		trimmed := strings.TrimSpace(*message)
		if trimmed == "" {
			message = nil
		} else {
			message = &trimmed
		}
	}
	chat := &model.Chat{RoomID: roomID, SenderID: senderID, Message: message}
	if err := s.chatRepo.Save(chat); err != nil {
		return nil, err
	}
	return chat, nil
}

func (s *ChatService) GetMessages(roomID, userID uint, page, limit int) ([]model.Chat, int64, error) {
	if !s.roomRepo.IsMember(roomID, userID) {
		return nil, 0, errors.New("forbidden")
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.chatRepo.FindByRoomID(roomID, page, limit)
}

func (s *ChatService) Unsend(ctx context.Context, chatID, senderID uint) error {
	chat, err := s.chatRepo.FindByID(chatID)
	if err != nil {
		return errors.New("chat not found")
	}
	if chat.SenderID != senderID {
		return errors.New("forbidden")
	}
	return s.chatRepo.SoftDelete(chatID)
}

func (s *ChatService) MarkRead(roomID, userID uint) error {
	if !s.roomRepo.IsMember(roomID, userID) {
		return errors.New("forbidden")
	}
	return s.chatRepo.MarkRoomRead(roomID, userID)
}
