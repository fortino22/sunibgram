package repository

type IChatReadRepository interface {
	MarkRead(chatID, userID uint) error
	HasRead(chatID, userID uint) bool
	CountReads(chatID uint) int64
}
