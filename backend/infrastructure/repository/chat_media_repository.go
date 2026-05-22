package repository

import (
	"gorm.io/gorm"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
)

type ChatMediaRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewChatMediaRepository(writeDB, readDB *gorm.DB) domainrepo.IChatMediaRepository {
	return &ChatMediaRepository{writeDB: writeDB, readDB: readDB}
}

func (r *ChatMediaRepository) Save(media *model.ChatMedia) error {
	return r.writeDB.Create(media).Error
}

func (r *ChatMediaRepository) FindByChatID(chatID uint) ([]model.ChatMedia, error) {
	var media []model.ChatMedia
	err := r.readDB.Where("chat_id = ?", chatID).Find(&media).Error
	return media, err
}

func (r *ChatMediaRepository) FindByChatIDs(chatIDs []uint) ([]model.ChatMedia, error) {
	var media []model.ChatMedia
	if len(chatIDs) == 0 {
		return media, nil
	}
	err := r.readDB.Where("chat_id IN ?", chatIDs).Find(&media).Error
	return media, err
}

func (r *ChatMediaRepository) DeleteByChatID(chatID uint) error {
	return r.writeDB.Where("chat_id = ?", chatID).Delete(&model.ChatMedia{}).Error
}
