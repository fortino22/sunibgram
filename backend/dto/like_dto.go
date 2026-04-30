package dto

import "github.com/sunibgram/dto/utils"

type LikedThreadResponse struct {
	ThreadResponse
	LikedAt utils.FormattedTime `json:"liked_at"`
}

type PaginatedLikedThreadsResponse struct {
	Data  []LikedThreadResponse `json:"data"`
	Total int64                 `json:"total"`
	Page  int                   `json:"page"`
	Limit int                   `json:"limit"`
}
