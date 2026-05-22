package repository

import (
	"gorm.io/gorm"

	chatmediamodel "github.com/sunibgram/domain/model/chat_media"
	chatmediarepoiface "github.com/sunibgram/domain/repository/chat_media"
)

type ChatMediaRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewChatMediaRepository(writeDB, readDB *gorm.DB) chatmediarepoiface.IChatMediaRepository {
	return &ChatMediaRepository{writeDB: writeDB, readDB: readDB}
}

func (r *ChatMediaRepository) Save(media *chatmediamodel.ChatMedia) error {
	return r.writeDB.Create(media).Error
}

func (r *ChatMediaRepository) FindByChatID(chatID uint) ([]chatmediamodel.ChatMedia, error) {
	var media []chatmediamodel.ChatMedia
	err := r.readDB.Where("chat_id = ?", chatID).Find(&media).Error
	return media, err
}

func (r *ChatMediaRepository) FindByChatIDs(chatIDs []uint) ([]chatmediamodel.ChatMedia, error) {
	var media []chatmediamodel.ChatMedia
	if len(chatIDs) == 0 {
		return media, nil
	}
	err := r.readDB.Where("chat_id IN ?", chatIDs).Find(&media).Error
	return media, err
}

func (r *ChatMediaRepository) DeleteByChatID(chatID uint) error {
	return r.writeDB.Where("chat_id = ?", chatID).Delete(&chatmediamodel.ChatMedia{}).Error
}
