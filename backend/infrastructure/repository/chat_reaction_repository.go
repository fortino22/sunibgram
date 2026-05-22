package repository

import (
	"gorm.io/gorm"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
)

type ChatReactionRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewChatReactionRepository(writeDB, readDB *gorm.DB) domainrepo.IChatReactionRepository {
	return &ChatReactionRepository{writeDB: writeDB, readDB: readDB}
}

func (r *ChatReactionRepository) Save(reaction *model.ChatReaction) error {
	return r.writeDB.
		Where(model.ChatReaction{ChatID: reaction.ChatID, UserID: reaction.UserID, Emoji: reaction.Emoji}).
		FirstOrCreate(reaction).Error
}

func (r *ChatReactionRepository) Delete(chatID, userID uint, emoji string) error {
	return r.writeDB.
		Where("chat_id = ? AND user_id = ? AND emoji = ?", chatID, userID, emoji).
		Delete(&model.ChatReaction{}).Error
}

func (r *ChatReactionRepository) FindByChatIDs(chatIDs []uint) ([]model.ChatReaction, error) {
	if len(chatIDs) == 0 {
		return []model.ChatReaction{}, nil
	}
	var reactions []model.ChatReaction
	err := r.readDB.Where("chat_id IN ?", chatIDs).Find(&reactions).Error
	return reactions, err
}
