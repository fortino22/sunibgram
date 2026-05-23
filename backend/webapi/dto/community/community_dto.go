package communitydto

import (
	utils "github.com/sunibgram/webapi/dto/_shared"
	communityruledto "github.com/sunibgram/webapi/dto/community_rule"

	threaddto "github.com/sunibgram/webapi/dto/thread"
)

type CommunityResponse struct {
	CommunityID  uint                                         `json:"community_id"`
	UserID       uint                                         `json:"user_id"`
	Username     string                                       `json:"username"`
	Name         string                                       `json:"name"`
	Description  string                                       `json:"description"`
	Rules        []communityruledto.CommunityRuleMiniResponse `json:"rules,omitempty"`
	CategoryName string                                       `json:"category_name"`
	Visibility   string                                       `json:"visibility"`
	BannerURL    string                                       `json:"banner_url"`
	MemberCount  int                                          `json:"member_count"`
	TopMembers   []CommunityTopMemberResponse                 `json:"top_members"`
	CreatedAt    utils.FormattedTime                          `json:"created_at"`
	UpdatedAt    utils.FormattedTime                          `json:"updated_at"`
}

type PaginatedCommunityResponse struct {
	Data []CommunityResponse `json:"data"`
	Meta threaddto.FeedMeta  `json:"meta"`
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

type CommunityAboutInfoResponse struct {
	Description      string `json:"description"`
	PostingPolicy    string `json:"posting_policy"`
	VisibilityPolicy string `json:"visibility_policy"`
	JoinPolicy       string `json:"join_policy"`
}

type CommunityAboutCreatorResponse struct {
	UserID      uint                `json:"user_id"`
	Name        string              `json:"name"`
	Username    string              `json:"username"`
	AvatarURL   string              `json:"avatar_url"`
	Bio         string              `json:"bio"`
	IsPremium   bool                `json:"is_premium"`
	IsFollowing bool                `json:"is_following"`
	CreatedAt   utils.FormattedTime `json:"community_created_at"`
}

type CommunityAboutMemberResponse struct {
	UserID      uint   `json:"user_id"`
	Name        string `json:"name"`
	Username    string `json:"username"`
	AvatarURL   string `json:"avatar_url"`
	Bio         string `json:"bio"`
	IsPremium   bool   `json:"is_premium"`
	IsFollowing bool   `json:"is_following"`
}

type CommunityAboutResponse struct {
	CommunityID uint                                         `json:"community_id"`
	Name        string                                       `json:"name"`
	Info        CommunityAboutInfoResponse                   `json:"info"`
	Creator     CommunityAboutCreatorResponse                `json:"creator"`
	Rules       []communityruledto.CommunityRuleMiniResponse `json:"rules"`
	Admins      []CommunityAboutMemberResponse               `json:"admins"`
	TopMembers  []CommunityAboutMemberResponse               `json:"top_members"`
}
