package dto

import "github.com/sunibgram/dto/utils"

type CommunityResponse struct {
	CommunityID  uint                         `json:"community_id"`
	UserID       uint                         `json:"user_id"`
	Username     string                       `json:"username"`
	Name         string                       `json:"name"`
	Description  string                       `json:"description"`
	Rules        string                       `json:"rules"`
	CategoryName string                       `json:"category_name"`
	Visibility   string                       `json:"visibility"`
	AvatarURL    string                       `json:"avatar_url"`
	BannerURL    string                       `json:"banner_url"`
	MemberCount  int                          `json:"member_count"`
	TopMembers   []CommunityTopMemberResponse `json:"top_members"`
	CreatedAt    utils.FormattedTime          `json:"created_at"`
	UpdatedAt    utils.FormattedTime          `json:"updated_at"`
}

type PaginatedCommunityResponse struct {
	Data []CommunityResponse `json:"data"`
	Meta FeedMeta            `json:"meta"`
}

type DiscoverCommunityMemberResponse struct {
	UserID    string `json:"user_id"`
	AvatarURL string `json:"avatar_url"`
}

type CommunityTopMemberResponse struct {
	UserID    string `json:"user_id"`
	AvatarURL string `json:"avatar_url"`
}

type DiscoverCommunityResponse struct {
	CommunityID  uint                              `json:"community_id"`
	Name         string                            `json:"name"`
	BannerURL    string                            `json:"banner_url"`
	MemberCount  int                               `json:"member_count"`
	CategoryName string                            `json:"category_name"`
	TopMembers   []DiscoverCommunityMemberResponse `json:"top_members"`
}

type MyCommunityResponse struct {
	CommunityID   uint   `json:"community_id"`
	CommunityName string `json:"community_name"`
	BannerURL     string `json:"community_banner"`
}
