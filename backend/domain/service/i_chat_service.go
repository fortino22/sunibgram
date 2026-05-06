package service

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type IChatService interface {
	Send(ctx context.Context, roomID, senderID uint, message *string) (*model.Chat, error)
	GetMessages(roomID, userID uint, page, limit int) ([]model.Chat, int64, error)
	Unsend(ctx context.Context, chatID, senderID uint) error
	MarkRead(roomID, userID uint) error
}
