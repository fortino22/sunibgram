package threadsummarydto

import (
	dtoutils "github.com/sunibgram/webapi/dto/_shared"
	threaddto "github.com/sunibgram/webapi/dto/thread"
	threadmediadto "github.com/sunibgram/webapi/dto/thread_media"
)

// SummaryHistoryItem is one row in a user's summary history: the thread content
// that was summarized plus the generated summary result.
type SummaryHistoryItem struct {
	ThreadID     uint                                 `json:"thread_id"`
	Content      string                               `json:"content"`
	Summary      string                               `json:"summary"`
	IsStale      bool                                 `json:"is_stale"`
	SummarizedAt dtoutils.FormattedTime               `json:"summarized_at"`
	Media        []threadmediadto.ThreadMediaResponse `json:"media"`
}

type PaginatedSummaryHistoryResponse struct {
	Data []SummaryHistoryItem `json:"data"`
	Meta threaddto.FeedMeta   `json:"meta"`
}
