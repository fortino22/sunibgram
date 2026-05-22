package repository

import (
	"time"

	"gorm.io/gorm"

	chatmodel "github.com/sunibgram/domain/model/chat"
	chatmediamodel "github.com/sunibgram/domain/model/chat_media"
	chatreadmodel "github.com/sunibgram/domain/model/chat_read"
	chatroommodel "github.com/sunibgram/domain/model/chat_room"
	chatrepoiface "github.com/sunibgram/domain/repository/chat"
)

type ChatRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewChatRepository(writeDB, readDB *gorm.DB) chatrepoiface.IChatRepository {
	return &ChatRepository{writeDB: writeDB, readDB: readDB}
}

func (r *ChatRepository) Save(chat *chatmodel.Chat) error {
	return r.writeDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(chat).Error; err != nil {
			return err
		}
		if err := tx.Model(&chatroommodel.ChatRoom{}).
			Where("room_id = ?", chat.RoomID).
			Update("updated_at", time.Now()).Error; err != nil {
			return err
		}
		return tx.Preload("Sender").Preload("ReplyTo.Sender").First(chat, chat.ChatID).Error
	})
}

func (r *ChatRepository) FindByRoomID(roomID uint, page, limit int) ([]chatmodel.Chat, int64, error) {
	var total int64
	if err := r.readDB.Model(&chatmodel.Chat{}).
		Where("room_id = ? AND deleted_at IS NULL", roomID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var chats []chatmodel.Chat
	offset := (page - 1) * limit
	err := r.readDB.Preload("Sender").Preload("ReplyTo.Sender").
		Where("room_id = ? AND deleted_at IS NULL", roomID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&chats).Error
	return chats, total, err
}

func (r *ChatRepository) FindByID(id uint) (*chatmodel.Chat, error) {
	var chat chatmodel.Chat
	err := r.readDB.Preload("Sender").Preload("ReplyTo.Sender").
		Where("deleted_at IS NULL").
		First(&chat, id).Error
	return &chat, err
}

func (r *ChatRepository) FindByIDs(ids []uint) ([]chatmodel.Chat, error) {
	if len(ids) == 0 {
		return []chatmodel.Chat{}, nil
	}
	var chats []chatmodel.Chat
	err := r.readDB.Preload("Sender").Preload("ReplyTo.Sender").
		Where("chat_id IN ? AND deleted_at IS NULL", ids).
		Find(&chats).Error
	return chats, err
}

func (r *ChatRepository) SoftDelete(id uint) error {
	now := time.Now()
	return r.writeDB.Model(&chatmodel.Chat{}).
		Where("chat_id = ?", id).
		Update("deleted_at", &now).Error
}

func (r *ChatRepository) Delete(id uint) error {
	return r.writeDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("chat_id = ?", id).Delete(&chatmediamodel.ChatMedia{}).Error; err != nil {
			return err
		}
		if err := tx.Where("chat_id = ?", id).Delete(&chatreadmodel.ChatRead{}).Error; err != nil {
			return err
		}
		return tx.Delete(&chatmodel.Chat{}, "chat_id = ?", id).Error
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
	return r.writeDB.Model(&chatmodel.Chat{}).
		Where("chat_id = ? AND deleted_at IS NULL", chatID).
		Updates(map[string]any{"message": message, "is_edited": true, "edited_at": &now}).Error
}

func (r *ChatRepository) UpdateIsRead(chatID uint) error {
	return r.writeDB.Model(&chatmodel.Chat{}).
		Where("chat_id = ? AND deleted_at IS NULL", chatID).
		Update("is_read", true).Error
}
