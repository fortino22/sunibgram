package dto

import "github.com/sunibgram/dto/utils"

type UserProfileResponse struct {
	UserID         uint                `json:"user_id"`
	Name           string              `json:"name"`
	Username       string              `json:"username"`
	Email          string              `json:"email"`
	FollowerCount  int                 `json:"follower_count"`
	FollowingCount int                 `json:"following_count"`
	CreatedAt      utils.FormattedTime `json:"created_at"`
}
