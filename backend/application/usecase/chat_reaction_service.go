package usecase

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type ChatReactionService struct {
	chatRepo     repository.IChatRepository
	reactionRepo repository.IChatReactionRepository
	roomRepo     repository.IChatRoomRepository
}

func NewChatReactionService(chatRepo repository.IChatRepository, reactionRepo repository.IChatReactionRepository, roomRepo repository.IChatRoomRepository) service.IChatReactionService {
	return &ChatReactionService{chatRepo: chatRepo, reactionRepo: reactionRepo, roomRepo: roomRepo}
}

func (s *ChatReactionService) Add(ctx context.Context, chatID, userID uint, emoji string) (*model.Chat, error) {
	emoji = strings.TrimSpace(emoji)
	if emoji == "" {
		return nil, errors.New("emoji is required")
	}
	if utf8.RuneCountInString(emoji) > 32 {
		return nil, errors.New("emoji must be 32 characters or less")
	}
	chat, err := s.chatRepo.FindByID(chatID)
	if err != nil {
		return nil, errors.New("chat not found")
	}
	if !s.roomRepo.IsMember(chat.RoomID, userID) {
		return nil, errors.New("forbidden")
	}
	reaction := &model.ChatReaction{ChatID: chatID, UserID: userID, Emoji: emoji}
	if err := s.reactionRepo.Save(reaction); err != nil {
		return nil, err
	}
	return chat, nil
}

func (s *ChatReactionService) Remove(ctx context.Context, chatID, userID uint, emoji string) (*model.Chat, error) {
	chat, err := s.chatRepo.FindByID(chatID)
	if err != nil {
		return nil, errors.New("chat not found")
	}
	if !s.roomRepo.IsMember(chat.RoomID, userID) {
		return nil, errors.New("forbidden")
	}
	if err := s.reactionRepo.Delete(chatID, userID, emoji); err != nil {
		return nil, err
	}
	return chat, nil
}

func (s *ChatReactionService) GetByChatIDs(chatIDs []uint) ([]model.ChatReaction, error) {
	return s.reactionRepo.FindByChatIDs(chatIDs)
}
