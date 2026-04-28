package dto

import "time"

type RepostedThreadResponse struct {
	ThreadResponse
	RepostedAt time.Time `json:"reposted_at"`
}

type PaginatedRepostedThreadsResponse struct {
	Data  []RepostedThreadResponse `json:"data"`
	Total int64                    `json:"total"`
	Page  int                      `json:"page"`
	Limit int                      `json:"limit"`
}
