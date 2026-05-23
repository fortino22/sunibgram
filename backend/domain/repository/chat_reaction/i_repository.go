package chatreactionrepoiface

import chatreactionmodel "github.com/sunibgram/domain/model/chat_reaction"

type IChatReactionRepository interface {
	Save(reaction *chatreactionmodel.ChatReaction) error
	Delete(chatID, userID uint, emoji string) error
	FindByChatIDs(chatIDs []uint) ([]chatreactionmodel.ChatReaction, error)
}
