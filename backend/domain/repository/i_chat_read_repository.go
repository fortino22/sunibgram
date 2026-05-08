package repository

import "context"

type IChatReadRepository interface {
	MarkRead(chatID, userID uint) error
	BatchMarkRead(roomID, userID uint) ([]uint, error)
	HasRead(chatID, userID uint) bool
	CountReads(chatID uint) int64
	CountReadsBatch(ctx context.Context, chatIDs []uint) (map[uint]int64, error)
}
