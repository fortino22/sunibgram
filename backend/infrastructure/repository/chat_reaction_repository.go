package repository

import (
	"gorm.io/gorm"

	chatreactionmodel "github.com/sunibgram/domain/model/chat_reaction"
	chatreactionrepoiface "github.com/sunibgram/domain/repository/chat_reaction"
)

type ChatReactionRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewChatReactionRepository(writeDB, readDB *gorm.DB) chatreactionrepoiface.IChatReactionRepository {
	return &ChatReactionRepository{writeDB: writeDB, readDB: readDB}
}

func (r *ChatReactionRepository) Save(reaction *chatreactionmodel.ChatReaction) error {
	return r.writeDB.
		Where(chatreactionmodel.ChatReaction{ChatID: reaction.ChatID, UserID: reaction.UserID, Emoji: reaction.Emoji}).
		FirstOrCreate(reaction).Error
}

func (r *ChatReactionRepository) Delete(chatID, userID uint, emoji string) error {
	return r.writeDB.
		Where("chat_id = ? AND user_id = ? AND emoji = ?", chatID, userID, emoji).
		Delete(&chatreactionmodel.ChatReaction{}).Error
}

func (r *ChatReactionRepository) FindByChatIDs(chatIDs []uint) ([]chatreactionmodel.ChatReaction, error) {
	if len(chatIDs) == 0 {
		return []chatreactionmodel.ChatReaction{}, nil
	}
	var reactions []chatreactionmodel.ChatReaction
	err := r.readDB.Where("chat_id IN ?", chatIDs).Find(&reactions).Error
	return reactions, err
}
