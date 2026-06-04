package foxchatsvciface

import (
	"context"
	"io"
)

// QAPair is one prior turn of the conversation, passed back by the client so
// the Fox can keep short-term context.
type QAPair struct {
	Question string
	Answer   string
}

type IFoxChatService interface {
	// ResolveChatSession returns the session to append this ask to: when
	// sessionID is 0 (or not owned by the user) it creates a new chat session
	// titled from firstQuestion; otherwise it returns the existing session ID.
	ResolveChatSession(ctx context.Context, userID, sessionID uint, firstQuestion string) (uint, error)

	// StreamAsk answers the user's question grounded in the LM wiki, streaming
	// the answer to w as it is generated. history carries recent prior turns
	// for context. On completion the turn is appended to sessionID. Unlimited —
	// no quota.
	StreamAsk(ctx context.Context, userID, sessionID uint, question string, history []QAPair, w io.Writer) error
}
