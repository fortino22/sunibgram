package repository

import "github.com/sunibgram/domain/model"

type IChatRepository interface {
	Save(chat *model.Chat) error
	FindByRoomID(roomID uint, page, limit int) ([]model.Chat, int64, error)
	FindByID(id uint) (*model.Chat, error)
	FindByIDs(ids []uint) ([]model.Chat, error)
	SoftDelete(id uint) error
	MarkRoomRead(roomID, userID uint) ([]uint, error)
	UpdateMessage(chatID uint, message string) error
	UpdateIsRead(chatID uint) error
}
