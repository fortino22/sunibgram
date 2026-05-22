package service

import (
	"context"

	chatmodel "github.com/sunibgram/domain/model/chat"
	chatreactionmodel "github.com/sunibgram/domain/model/chat_reaction"
)

type IChatReactionService interface {
	Add(ctx context.Context, chatID, userID uint, emoji string) (*chatmodel.Chat, error)
	Remove(ctx context.Context, chatID, userID uint, emoji string) (*chatmodel.Chat, error)
	GetByChatIDs(chatIDs []uint) ([]chatreactionmodel.ChatReaction, error)
}
