package dto

import "github.com/sunibgram/dto/utils"

type RepostedThreadResponse struct {
	ThreadResponse
	RepostedAt utils.FormattedTime `json:"reposted_at"`
}

type PaginatedRepostedThreadsResponse struct {
	Data  []RepostedThreadResponse `json:"data"`
	Total int64                    `json:"total"`
	Page  int                      `json:"page"`
	Limit int                      `json:"limit"`
}
