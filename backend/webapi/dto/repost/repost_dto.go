package repostdto

import (
	utils "github.com/sunibgram/webapi/dto/_shared"
	threaddto "github.com/sunibgram/webapi/dto/thread"
)

type RepostedThreadResponse struct {
	threaddto.ThreadResponse
	RepostedAt utils.FormattedTime `json:"reposted_at"`
}

type PaginatedRepostedThreadsResponse struct {
	Data  []RepostedThreadResponse `json:"data"`
	Total int64                    `json:"total"`
	Page  int                      `json:"page"`
	Limit int                      `json:"limit"`
}
