package repository

import (
	"time"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ChatReadRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewChatReadRepository(writeDB, readDB *gorm.DB) domainrepo.IChatReadRepository {
	return &ChatReadRepository{writeDB: writeDB, readDB: readDB}
}

func (r *ChatReadRepository) MarkRead(chatID, userID uint) error {
	read := model.ChatRead{ChatID: chatID, UserID: userID, ReadAt: time.Now()}
	return r.writeDB.Clauses(clause.OnConflict{DoNothing: true}).Create(&read).Error
}

func (r *ChatReadRepository) BatchMarkRead(roomID, userID uint) ([]uint, error) {
	var chatIDs []uint
	err := r.writeDB.Raw(`
		INSERT INTO chat_reads (chat_id, user_id, read_at)
		SELECT chat_id, ?, NOW()
		FROM chats
		WHERE room_id = ?
			AND sender_id != ?
			AND deleted_at IS NULL
		ON CONFLICT (chat_id, user_id) DO NOTHING
		RETURNING chat_id
	`, userID, roomID, userID).Scan(&chatIDs).Error
	return chatIDs, err
}

func (r *ChatReadRepository) HasRead(chatID, userID uint) bool {
	var count int64
	if err := r.readDB.Model(&model.ChatRead{}).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Count(&count).Error; err != nil {
		return false
	}
	return count > 0
}

func (r *ChatReadRepository) CountReads(chatID uint) int64 {
	var count int64
	if err := r.readDB.Table("chat_reads").
		Where("chat_id = ?", chatID).
		Count(&count).Error; err != nil {
		return 0
	}
	return count
}
