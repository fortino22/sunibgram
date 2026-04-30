package dto

import "github.com/sunibgram/dto/utils"

type BookmarkedThreadResponse struct {
	ThreadResponse
	BookmarkedAt utils.FormattedTime `json:"bookmarked_at"`
}

type PaginatedBookmarkedThreadsResponse struct {
	Data  []BookmarkedThreadResponse `json:"data"`
	Total int64                      `json:"total"`
	Page  int                        `json:"page"`
	Limit int                        `json:"limit"`
}
