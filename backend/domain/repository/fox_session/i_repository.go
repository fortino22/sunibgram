package foxsessionrepoiface

import (
	foxhistorymodel "github.com/sunibgram/domain/model/fox_history"
	foxsessionmodel "github.com/sunibgram/domain/model/fox_session"
)

// ChatTurn is one question/answer pair in a chat session, oldest-first.
type ChatTurn struct {
	Question string
	Answer   string
}

// SessionPreview is a session plus the fields the history list needs to render a
// row without a second round-trip. Entry is the representative fox_histories
// row: for summary sessions it is the single summary entry (Body/Snapshot); for
// chat sessions it is the latest turn (may be nil if the session has no turns
// yet). TurnCount is the number of chat turns (0 for summary sessions).
type SessionPreview struct {
	Session   foxsessionmodel.FoxSession
	Entry     *foxhistorymodel.FoxHistory
	TurnCount int
}

type IFoxSessionRepository interface {
	// CreateChatSession starts a new empty chat session and returns its ID.
	CreateChatSession(userID uint, title string) (uint, error)

	// AppendChatTurn inserts one fox_histories chat entry for the session and
	// touches the session's UpdatedAt so it sorts to the top of the history.
	AppendChatTurn(sessionID uint, question, answer string) error

	// UpsertSummarySession creates or refreshes the summary session for
	// (userID, threadID) and its single summary entry. On conflict it refreshes
	// the summary/snapshot and touches UpdatedAt. Returns the session ID.
	UpsertSummarySession(userID, threadID uint, summary, snapshot string) (uint, error)

	// OwnsSession reports whether the session exists and belongs to the user.
	OwnsSession(userID, sessionID uint) (bool, error)

	// GetSessionTurns returns a chat session's turns oldest-first. Returns an
	// empty slice for summary sessions or sessions the user does not own.
	GetSessionTurns(userID, sessionID uint) ([]ChatTurn, error)

	// ListSessions returns the user's sessions newest-first (by UpdatedAt) with
	// preview data, plus the total count. q filters (case-insensitive) on the
	// session title and on any entry's question/body/snapshot.
	ListSessions(userID uint, offset, limit int, q string) ([]SessionPreview, int64, error)
}
