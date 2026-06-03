package threadsummarydto

import (
	dtoutils "github.com/sunibgram/webapi/dto/_shared"
	threaddto "github.com/sunibgram/webapi/dto/thread"
	threadmediadto "github.com/sunibgram/webapi/dto/thread_media"
)

// SummaryHistoryItem is one row in a user's Fox history. Two kinds, by Type:
//   - "summary": ThreadID/Content/Summary/Media set (a thread the user summarized).
//   - "chat":    Question set + Summary = the Fox answer; ThreadID 0, no media.
type SummaryHistoryItem struct {
	ID           uint                                 `json:"id"`
	Type         string                               `json:"type"`
	ThreadID     uint                                 `json:"thread_id"`
	Content      string                               `json:"content"`
	Question     string                               `json:"question"`
	Summary      string                               `json:"summary"`
	IsStale      bool                                 `json:"is_stale"`
	SummarizedAt dtoutils.FormattedTime               `json:"summarized_at"`
	Media        []threadmediadto.ThreadMediaResponse `json:"media"`
}

type PaginatedSummaryHistoryResponse struct {
	Data []SummaryHistoryItem `json:"data"`
	Meta threaddto.FeedMeta   `json:"meta"`
}
