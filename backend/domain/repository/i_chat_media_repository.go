package repository

import "github.com/sunibgram/domain/model"

type IChatMediaRepository interface {
	Save(media *model.ChatMedia) error
	FindByChatID(chatID uint) ([]model.ChatMedia, error)
	FindByChatIDs(chatIDs []uint) ([]model.ChatMedia, error)
	DeleteByChatID(chatID uint) error
}
