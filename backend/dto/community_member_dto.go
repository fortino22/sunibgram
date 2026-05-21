package dto

import "github.com/sunibgram/dto/utils"

type CommunityMemberResponse struct {
	CommunityID uint                `json:"community_id"`
	UserID      uint                `json:"user_id"`
	Username    string              `json:"username"`
	AvatarURL   string              `json:"avatar_url"`
	Role        string              `json:"role"`
	JoinedAt    utils.FormattedTime `json:"joined_at"`
}

type PaginatedCommunityMemberResponse struct {
	Data []CommunityMemberResponse `json:"data"`
	Meta FeedMeta                  `json:"meta"`
}

type CommunityUserResponse struct {
	UserID      uint   `json:"user_id"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	IsPremium   bool   `json:"is_premium"`
	Role        string `json:"role"`
	IsFollowing bool   `json:"is_following"`
}

type PaginatedCommunityUserResponse struct {
	Data []CommunityUserResponse `json:"data"`
	Meta FeedMeta                `json:"meta"`
}
