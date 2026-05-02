package dto

import "time"

type CommunityResponse struct {
	CommunityID uint      `json:"community_id"`
	UserID      uint      `json:"user_id"`
	Username    string    `json:"username"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Rules       string    `json:"rules"`
	Category    string    `json:"category"`
	Visibility  string    `json:"visibility"`
	AvatarURL   string    `json:"avatar_url"`
	BannerURL   string    `json:"banner_url"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PaginatedCommunityResponse struct {
	Data []CommunityResponse `json:"data"`
	Meta FeedMeta            `json:"meta"`
}
