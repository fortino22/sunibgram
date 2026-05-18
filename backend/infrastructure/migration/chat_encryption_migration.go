package migration

import (
	"log"
	"strings"

	"github.com/sunibgram/application/helper"
	"gorm.io/gorm"
)

func migrateChatEncryption(db *gorm.DB) {
	db.Exec("ALTER TABLE chat_rooms ADD COLUMN IF NOT EXISTS encrypted_key TEXT")

	var roomIDs []uint
	if err := db.Table("chat_rooms").
		Where("encrypted_key IS NULL OR encrypted_key = ''").
		Pluck("room_id", &roomIDs).Error; err != nil {
		log.Fatalf("load chat rooms missing encryption keys: %v", err)
	}
	if len(roomIDs) == 0 {
		return
	}

	masterKey, err := helper.ChatMasterKeyFromEnv("CHAT_MASTER_KEY")
	if err != nil {
		log.Fatalf("CHAT_MASTER_KEY is required to encrypt existing chat rooms: %v", err)
	}

	for _, roomID := range roomIDs {
		if err := encryptExistingRoomMessages(db, masterKey, roomID); err != nil {
			log.Fatalf("encrypt existing chat room %d: %v", roomID, err)
		}
	}
}

func encryptExistingRoomMessages(db *gorm.DB, masterKey []byte, roomID uint) error {
	type chatRow struct {
		ChatID  uint
		Message *string
	}

	return db.Transaction(func(tx *gorm.DB) error {
		roomKey, err := helper.ChatGenerateKey()
		if err != nil {
			return err
		}
		encryptedKey, err := helper.ChatEncrypt(roomKey, masterKey)
		if err != nil {
			return err
		}
		if err := tx.Table("chat_rooms").
			Where("room_id = ?", roomID).
			Update("encrypted_key", encryptedKey).Error; err != nil {
			return err
		}

		var chats []chatRow
		if err := tx.Table("chats").
			Select("chat_id, message").
			Where("room_id = ? AND message IS NOT NULL AND message != ''", roomID).
			Find(&chats).Error; err != nil {
			return err
		}
		for _, chat := range chats {
			if chat.Message == nil || strings.TrimSpace(*chat.Message) == "" {
				continue
			}
			encryptedMessage, err := helper.ChatEncrypt([]byte(*chat.Message), roomKey)
			if err != nil {
				return err
			}
			if err := tx.Table("chats").
				Where("chat_id = ?", chat.ChatID).
				Update("message", encryptedMessage).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
