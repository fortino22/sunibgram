package service

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type IChatReactionService interface {
	Add(ctx context.Context, chatID, userID uint, emoji string) (*model.Chat, error)
	Remove(ctx context.Context, chatID, userID uint, emoji string) (*model.Chat, error)
	GetByChatIDs(chatIDs []uint) ([]model.ChatReaction, error)
}
