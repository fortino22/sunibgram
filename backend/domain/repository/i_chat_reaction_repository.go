package repository

import "github.com/sunibgram/domain/model"

type IChatReactionRepository interface {
	Save(reaction *model.ChatReaction) error
	Delete(chatID, userID uint, emoji string) error
	FindByChatIDs(chatIDs []uint) ([]model.ChatReaction, error)
}
