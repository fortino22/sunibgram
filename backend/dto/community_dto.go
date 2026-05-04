package dto

import "github.com/sunibgram/dto/utils"

type CommunityResponse struct {
	CommunityID uint                `json:"community_id"`
	UserID      uint                `json:"user_id"`
	Username    string              `json:"username"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Rules       string              `json:"rules"`
	Category    string              `json:"category"`
	Visibility  string              `json:"visibility"`
	AvatarURL   string              `json:"avatar_url"`
	BannerURL   string              `json:"banner_url"`
	MemberCount int                 `json:"member_count"`
	CreatedAt   utils.FormattedTime `json:"created_at"`
	UpdatedAt   utils.FormattedTime `json:"updated_at"`
}

type PaginatedCommunityResponse struct {
	Data []CommunityResponse `json:"data"`
	Meta FeedMeta            `json:"meta"`
}
