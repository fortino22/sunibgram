package communityrequestdto

import (
	utils "github.com/sunibgram/webapi/dto/_shared"
	threaddto "github.com/sunibgram/webapi/dto/thread"
)

type CommunityRequestResponse struct {
	CommunityID uint                `json:"community_id"`
	UserID      uint                `json:"user_id"`
	Name        string              `json:"name"`
	Username    string              `json:"username"`
	AvatarURL   string              `json:"avatar_url"`
	IsPremium   bool                `json:"is_premium"`
	Status      string              `json:"status"`
	CreatedAt   utils.FormattedTime `json:"created_at"`
}

type PaginatedCommunityRequestResponse struct {
	Data []CommunityRequestResponse `json:"data"`
	Meta threaddto.FeedMeta         `json:"meta"`
}
