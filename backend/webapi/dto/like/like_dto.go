package likedto

import (
	utils "github.com/sunibgram/webapi/dto/_shared"
	threaddto "github.com/sunibgram/webapi/dto/thread"
)

type LikedThreadResponse struct {
	threaddto.ThreadResponse
	LikedAt utils.FormattedTime `json:"liked_at"`
}

type PaginatedLikedThreadsResponse struct {
	Data  []LikedThreadResponse `json:"data"`
	Total int64                 `json:"total"`
	Page  int                   `json:"page"`
	Limit int                   `json:"limit"`
}
