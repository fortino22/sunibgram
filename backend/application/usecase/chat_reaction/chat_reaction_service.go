package chatreactionuc

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
	userblockrepoiface "github.com/sunibgram/domain/repository/user_block"
	chatreactionsvciface "github.com/sunibgram/domain/service/chat_reaction"
)

type ChatReactionService struct {
	chatRepo     chatrepoiface.IChatRepository
	reactionRepo chatreactionrepoiface.IChatReactionRepository
	roomRepo     chatroomrepoiface.IChatRoomRepository
	blockRepo    userblockrepoiface.IUserBlockRepository
}

func NewChatReactionService(chatRepo chatrepoiface.IChatRepository, reactionRepo chatreactionrepoiface.IChatReactionRepository, roomRepo chatroomrepoiface.IChatRoomRepository, blockRepo userblockrepoiface.IUserBlockRepository) chatreactionsvciface.IChatReactionService {
	return &ChatReactionService{chatRepo: chatRepo, reactionRepo: reactionRepo, roomRepo: roomRepo, blockRepo: blockRepo}
}

// privateRoomBlocked reports whether roomID is a private room whose other member
// is mutually blocked with userID. Group rooms return false.
func (s *ChatReactionService) privateRoomBlocked(roomID, userID uint) bool {
	if s.blockRepo == nil {
		return false
	}
	memberMap, err := s.roomRepo.FindOtherMemberInfo([]uint{roomID}, userID)
	if err != nil {
		return false
	}
	info, ok := memberMap[roomID]
	if !ok {
		return false
	}
	blocked, err := s.blockRepo.IsBlocked(userID, info.UserID)
	return err == nil && blocked
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
	if s.privateRoomBlocked(chat.RoomID, userID) {
		return nil, errors.New("blocked")
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
	if s.privateRoomBlocked(chat.RoomID, userID) {
		return nil, errors.New("blocked")
	}
	if err := s.reactionRepo.Delete(chatID, userID, emoji); err != nil {
		return nil, err
	}
	return chat, nil
}

func (s *ChatReactionService) GetByChatIDs(chatIDs []uint) ([]chatreactionmodel.ChatReaction, error) {
	return s.reactionRepo.FindByChatIDs(chatIDs)
}
