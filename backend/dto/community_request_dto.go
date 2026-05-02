package dto

import "time"

type CommunityRequestResponse struct {
	CommunityID uint      `json:"community_id"`
	UserID      uint      `json:"user_id"`
	Username    string    `json:"username"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type PaginatedCommunityRequestResponse struct {
	Data []CommunityRequestResponse `json:"data"`
	Meta FeedMeta                   `json:"meta"`
}
