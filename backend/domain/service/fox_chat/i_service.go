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
	// StreamAsk answers the user's question grounded in the LM wiki, streaming
	// the answer to w as it is generated. history carries recent prior turns
	// for context. Unlimited — no quota.
	StreamAsk(ctx context.Context, userID uint, question string, history []QAPair, w io.Writer) error
}
