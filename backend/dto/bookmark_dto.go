package dto

import "time"

type BookmarkedThreadResponse struct {
	ThreadResponse
	BookmarkedAt time.Time `json:"bookmarked_at"`
}

type PaginatedBookmarkedThreadsResponse struct {
	Data  []BookmarkedThreadResponse `json:"data"`
	Total int64                      `json:"total"`
	Page  int                        `json:"page"`
	Limit int                        `json:"limit"`
}
