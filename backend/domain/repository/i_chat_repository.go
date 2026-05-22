package repository

import chatmodel "github.com/sunibgram/domain/model/chat"

type IChatRepository interface {
	Save(chat *chatmodel.Chat) error
	FindByRoomID(roomID uint, page, limit int) ([]chatmodel.Chat, int64, error)
	FindByID(id uint) (*chatmodel.Chat, error)
	FindByIDs(ids []uint) ([]chatmodel.Chat, error)
	SoftDelete(id uint) error
	Delete(id uint) error
	MarkRoomRead(roomID, userID uint) ([]uint, error)
	UpdateMessage(chatID uint, message string) error
	UpdateIsRead(chatID uint) error
}
