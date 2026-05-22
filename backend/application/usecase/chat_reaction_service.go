package usecase

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	chatmodel "github.com/sunibgram/domain/model/chat"
	chatreactionmodel "github.com/sunibgram/domain/model/chat_reaction"
	chatrepoiface "github.com/sunibgram/domain/repository/chat"
	chatreactionrepoiface "github.com/sunibgram/domain/repository/chat_reaction"
	chatroomrepoiface "github.com/sunibgram/domain/repository/chat_room"
	chatreactionsvciface "github.com/sunibgram/domain/service/chat_reaction"
)

type ChatReactionService struct {
	chatRepo     chatrepoiface.IChatRepository
	reactionRepo chatreactionrepoiface.IChatReactionRepository
	roomRepo     chatroomrepoiface.IChatRoomRepository
}

func NewChatReactionService(chatRepo chatrepoiface.IChatRepository, reactionRepo chatreactionrepoiface.IChatReactionRepository, roomRepo chatroomrepoiface.IChatRoomRepository) chatreactionsvciface.IChatReactionService {
	return &ChatReactionService{chatRepo: chatRepo, reactionRepo: reactionRepo, roomRepo: roomRepo}
}

func (s *ChatReactionService) Add(ctx context.Context, chatID, userID uint, emoji string) (*chatmodel.Chat, error) {
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
	reaction := &chatreactionmodel.ChatReaction{ChatID: chatID, UserID: userID, Emoji: emoji}
	if err := s.reactionRepo.Save(reaction); err != nil {
		return nil, err
	}
	return chat, nil
}

func (s *ChatReactionService) Remove(ctx context.Context, chatID, userID uint, emoji string) (*chatmodel.Chat, error) {
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

func (s *ChatReactionService) GetByChatIDs(chatIDs []uint) ([]chatreactionmodel.ChatReaction, error) {
	return s.reactionRepo.FindByChatIDs(chatIDs)
}
