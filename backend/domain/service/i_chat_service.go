package service

import (
	"context"

	chatmodel "github.com/sunibgram/domain/model/chat"
)

type ChatMarkResult struct {
	ChatID   uint
	SenderID uint
	Status   string // "marked", "already_read", "not_found", "skipped"
}

type IChatService interface {
	Send(ctx context.Context, roomID, senderID uint, message *string, gifURL *string, replyToID *uint) (*chatmodel.Chat, error)
	GetByID(ctx context.Context, chatID uint) (*chatmodel.Chat, error)
	GetMessages(roomID, userID uint, page, limit int) ([]chatmodel.Chat, int64, error)
	Unsend(ctx context.Context, chatID, senderID uint) error
	Delete(ctx context.Context, chatID uint) error
	UpdateMessage(ctx context.Context, chatID, senderID uint, message string) error
	MarkRead(roomID, userID uint) ([]uint, error)
	MarkMessagesRead(ctx context.Context, roomID uint, chatIDs []uint, userID uint) ([]ChatMarkResult, error)
}
