package userprofiledto

import (
	utils "github.com/sunibgram/webapi/dto/_shared"
	threaddto "github.com/sunibgram/webapi/dto/thread"
)

type UserProfileResponse struct {
	UserID              uint                         `json:"user_id"`
	Name                string                       `json:"name"`
	Username            string                       `json:"username"`
	Email               string                       `json:"email"`
	IsPremium           bool                         `json:"is_premium"`
	Gender              string                       `json:"gender"`
	BirthdayDate        string                       `json:"birthday_date"`
	Bio                 string                       `json:"bio"`
	AvatarURL           string                       `json:"avatar_url"`
	BannerURL           string                       `json:"banner_url"`
	FollowerCount       int                          `json:"follower_count"`
	FollowingCount      int                          `json:"following_count"`
	IsFollowing         bool                         `json:"is_following"`
	IsBlocked           bool                         `json:"is_blocked"`
	IsBlockedBy         bool                         `json:"is_blocked_by"`
	IsMuted             bool                         `json:"is_muted"`
	OnboardingCompleted bool                         `json:"onboarding_completed"`
	Faculty             *threaddto.FeedAuthorFaculty `json:"faculty,omitempty"`
	CreatedAt           utils.FormattedTime          `json:"created_at"`
}

// OnboardingStatusResponse tells the frontend how many people the first-time
// user must follow (capped at the followable pool) and how many they already do.
type OnboardingStatusResponse struct {
	RequiredFollows int `json:"required_follows"`
	FollowingCount  int `json:"following_count"`
}

type UpdateProfileRequest struct {
	Name         string `json:"name,omitempty"`
	Username     string `json:"username,omitempty"`
	BirthdayDate string `json:"birthday_date,omitempty"`
	Bio          string `json:"bio,omitempty"`
}
