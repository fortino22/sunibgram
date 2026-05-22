package communitymemberdto

import (
	utils "github.com/sunibgram/webapi/dto/_shared"
	threaddto "github.com/sunibgram/webapi/dto/thread"
)

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
	Meta threaddto.FeedMeta        `json:"meta"`
}

type CommunityUserResponse struct {
	UserID      uint   `json:"user_id"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	UserAvatar  string `json:"user_avatar"`
	IsPremium   bool   `json:"is_premium"`
	Role        string `json:"role"`
	IsFollowing bool   `json:"is_following"`
}

type PaginatedCommunityUserResponse struct {
	Data []CommunityUserResponse `json:"data"`
	Meta threaddto.FeedMeta      `json:"meta"`
}
