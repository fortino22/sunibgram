package repository

type IChatReadRepository interface {
	MarkRead(chatID, userID uint) error
	BatchMarkRead(roomID, userID uint) ([]uint, error)
	HasRead(chatID, userID uint) bool
	CountReads(chatID uint) int64
}
