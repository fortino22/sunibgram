package dto

import "github.com/sunibgram/dto/utils"

type UserProfileResponse struct {
	UserID         uint                `json:"user_id"`
	Name           string              `json:"name"`
	Username       string              `json:"username"`
	Email          string              `json:"email"`
	IsVerified     bool                `json:"is_verified"`
	Bio            string              `json:"bio"`
	AvatarURL      string              `json:"avatar_url"`
	BannerURL      string              `json:"banner_url"`
	FollowerCount  int                 `json:"follower_count"`
	FollowingCount int                 `json:"following_count"`
	Faculty        *FeedAuthorFaculty  `json:"faculty,omitempty"`
	CreatedAt      utils.FormattedTime `json:"created_at"`
}
