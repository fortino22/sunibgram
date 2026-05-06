package service

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type ChatMarkResult struct {
	ChatID  uint
	Status  string // "marked", "already_read", "not_found", "skipped"
}

type IChatService interface {
	Send(ctx context.Context, roomID, senderID uint, message *string) (*model.Chat, error)
	GetMessages(roomID, userID uint, page, limit int) ([]model.Chat, int64, error)
	Unsend(ctx context.Context, chatID, senderID uint) error
	UpdateMessage(ctx context.Context, chatID, senderID uint, message string) error
	MarkRead(roomID, userID uint) error
	MarkMessagesRead(ctx context.Context, roomID uint, chatIDs []uint, userID uint) ([]ChatMarkResult, error)
}
