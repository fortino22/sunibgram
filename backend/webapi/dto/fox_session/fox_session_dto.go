package foxsessiondto

import (
	dtoutils "github.com/sunibgram/webapi/dto/_shared"
	threaddto "github.com/sunibgram/webapi/dto/thread"
	threadmediadto "github.com/sunibgram/webapi/dto/thread_media"
)

// SessionListItem is one row in the user's Fox history — one session. Two kinds
// by Type:
//   - "chat":    Title = first question, MessageCount = turn count, Preview =
//     latest answer snippet. Re-openable to continue (fetch turns via detail).
//   - "summary": ThreadID/Content/Summary/Media set — a thread the user
//     summarized (read-only).
type SessionListItem struct {
	ID           uint                                 `json:"id"`
	Type         string                               `json:"type"`
	ThreadID     uint                                 `json:"thread_id"`
	Title        string                               `json:"title"`
	Content      string                               `json:"content"`
	Summary      string                               `json:"summary"`
	Preview      string                               `json:"preview"`
	MessageCount int                                  `json:"message_count"`
	IsStale      bool                                 `json:"is_stale"`
	UpdatedAt    dtoutils.FormattedTime               `json:"updated_at"`
	Media        []threadmediadto.ThreadMediaResponse `json:"media"`
}

type PaginatedSessionsResponse struct {
	Data []SessionListItem  `json:"data"`
	Meta threaddto.FeedMeta `json:"meta"`
}

// SessionTurn is one question/answer pair in a chat session, oldest-first.
type SessionTurn struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// SessionDetailResponse returns a chat session's turns so the client can hydrate
// the Ask Fox panel and continue the conversation.
type SessionDetailResponse struct {
	ID    uint          `json:"id"`
	Turns []SessionTurn `json:"turns"`
}
