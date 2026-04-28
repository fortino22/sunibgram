package dto

import "time"

type LikedThreadResponse struct {
	ThreadResponse
	LikedAt time.Time `json:"liked_at"`
}

type PaginatedLikedThreadsResponse struct {
	Data  []LikedThreadResponse `json:"data"`
	Total int64                 `json:"total"`
	Page  int                   `json:"page"`
	Limit int                   `json:"limit"`
}
