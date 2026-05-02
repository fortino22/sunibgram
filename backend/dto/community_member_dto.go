package dto

import "time"

type CommunityMemberResponse struct {
	CommunityID uint      `json:"community_id"`
	UserID      uint      `json:"user_id"`
	Username    string    `json:"username"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}

type PaginatedCommunityMemberResponse struct {
	Data []CommunityMemberResponse `json:"data"`
	Meta FeedMeta                  `json:"meta"`
}
