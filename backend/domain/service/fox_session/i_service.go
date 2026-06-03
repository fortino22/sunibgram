package foxsessionsvciface

import (
	"context"

	foxsessionrepoiface "github.com/sunibgram/domain/repository/fox_session"
)

type IFoxSessionService interface {
	// ListSessions returns the user's Fox sessions newest-first (chats +
	// summaries) with preview data, plus the total count. q filters on the
	// session title and any entry's text.
	ListSessions(ctx context.Context, userID uint, page, limit int, q string) ([]foxsessionrepoiface.SessionPreview, int64, error)

	// GetSessionTurns returns a chat session's turns oldest-first (empty for
	// summary sessions or sessions the user does not own).
	GetSessionTurns(ctx context.Context, userID, sessionID uint) ([]foxsessionrepoiface.ChatTurn, error)
}
