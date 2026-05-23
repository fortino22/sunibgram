package bookmarkdto

import (
	utils "github.com/sunibgram/webapi/dto/_shared"
	threaddto "github.com/sunibgram/webapi/dto/thread"
)

type BookmarkedThreadResponse struct {
	threaddto.ThreadResponse
	BookmarkedAt utils.FormattedTime `json:"bookmarked_at"`
}

type PaginatedBookmarkedThreadsResponse struct {
	Data  []BookmarkedThreadResponse `json:"data"`
	Total int64                      `json:"total"`
	Page  int                        `json:"page"`
	Limit int                        `json:"limit"`
}
