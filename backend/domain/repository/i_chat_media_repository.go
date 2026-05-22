package repository

import chatmediamodel "github.com/sunibgram/domain/model/chat_media"

type IChatMediaRepository interface {
	Save(media *chatmediamodel.ChatMedia) error
	FindByChatID(chatID uint) ([]chatmediamodel.ChatMedia, error)
	FindByChatIDs(chatIDs []uint) ([]chatmediamodel.ChatMedia, error)
	DeleteByChatID(chatID uint) error
}
