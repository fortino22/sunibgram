package repository

import (
	"time"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"gorm.io/gorm"
)

type ChatRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewChatRepository(writeDB, readDB *gorm.DB) domainrepo.IChatRepository {
	return &ChatRepository{writeDB: writeDB, readDB: readDB}
}

func (r *ChatRepository) Save(chat *model.Chat) error {
	return r.writeDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(chat).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.ChatRoom{}).
			Where("room_id = ?", chat.RoomID).
			Update("updated_at", time.Now()).Error; err != nil {
			return err
		}
		return tx.Preload("Sender").Preload("ReplyTo.Sender").First(chat, chat.ChatID).Error
	})
}

func (r *ChatRepository) FindByRoomID(roomID uint, page, limit int) ([]model.Chat, int64, error) {
	var total int64
	if err := r.readDB.Model(&model.Chat{}).
		Where("room_id = ? AND deleted_at IS NULL", roomID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var chats []model.Chat
	offset := (page - 1) * limit
	err := r.readDB.Preload("Sender").Preload("ReplyTo.Sender").
		Where("room_id = ? AND deleted_at IS NULL", roomID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&chats).Error
	return chats, total, err
}

func (r *ChatRepository) FindByID(id uint) (*model.Chat, error) {
	var chat model.Chat
	err := r.readDB.Preload("Sender").Preload("ReplyTo.Sender").
		Where("deleted_at IS NULL").
		First(&chat, id).Error
	return &chat, err
}

func (r *ChatRepository) FindByIDs(ids []uint) ([]model.Chat, error) {
	if len(ids) == 0 {
		return []model.Chat{}, nil
	}
	var chats []model.Chat
	err := r.readDB.Preload("Sender").Preload("ReplyTo.Sender").
		Where("chat_id IN ? AND deleted_at IS NULL", ids).
		Find(&chats).Error
	return chats, err
}

func (r *ChatRepository) SoftDelete(id uint) error {
	now := time.Now()
	return r.writeDB.Model(&model.Chat{}).
		Where("chat_id = ?", id).
		Update("deleted_at", &now).Error
}

func (r *ChatRepository) Delete(id uint) error {
	return r.writeDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("chat_id = ?", id).Delete(&model.ChatMedia{}).Error; err != nil {
			return err
		}
		if err := tx.Where("chat_id = ?", id).Delete(&model.ChatRead{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Chat{}, "chat_id = ?", id).Error
	})
}

func (r *ChatRepository) MarkRoomRead(roomID, userID uint) ([]uint, error) {
	var ids []uint
	err := r.writeDB.Raw(
		`UPDATE chats SET is_read = true WHERE room_id = ? AND sender_id != ? AND is_read = false AND deleted_at IS NULL RETURNING chat_id`,
		roomID, userID,
	).Scan(&ids).Error
	return ids, err
}

func (r *ChatRepository) UpdateMessage(chatID uint, message string) error {
	now := time.Now()
	return r.writeDB.Model(&model.Chat{}).
		Where("chat_id = ? AND deleted_at IS NULL", chatID).
		Updates(map[string]any{"message": message, "is_edited": true, "edited_at": &now}).Error
}

func (r *ChatRepository) UpdateIsRead(chatID uint) error {
	return r.writeDB.Model(&model.Chat{}).
		Where("chat_id = ? AND deleted_at IS NULL", chatID).
		Update("is_read", true).Error
}
