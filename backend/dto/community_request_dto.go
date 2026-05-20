package dto

import "github.com/sunibgram/dto/utils"

type CommunityRequestResponse struct {
	CommunityID uint                `json:"community_id"`
	UserID      uint                `json:"user_id"`
	Username    string              `json:"username"`
	UserAvatar  string              `json:"user_avatar"`
	IsPremium   bool                `json:"is_premium"`
	Status      string              `json:"status"`
	CreatedAt   utils.FormattedTime `json:"created_at"`
}

type PaginatedCommunityRequestResponse struct {
	Data []CommunityRequestResponse `json:"data"`
	Meta FeedMeta                   `json:"meta"`
}
