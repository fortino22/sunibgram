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
	FacultyLocation     *FacultyLocationMini         `json:"faculty_location,omitempty"`
	CreatedAt           utils.FormattedTime          `json:"created_at"`
	// *ChangeAvailableAt are set only on self-view, and only while the field is
	// within its change cooldown — the edit UI uses them to disable the field
	// and show when the next change is allowed. Omitted otherwise.
	NameChangeAvailableAt     *utils.FormattedTime `json:"name_change_available_at,omitempty"`
	UsernameChangeAvailableAt *utils.FormattedTime `json:"username_change_available_at,omitempty"`
	// FacultyLocationChangeAvailableAt is self-view only: the campus cooldown
	// (~6 months) the edit UI uses to lock the field until the next allowed date.
	FacultyLocationChangeAvailableAt *utils.FormattedTime `json:"faculty_location_change_available_at,omitempty"`
}

// FacultyLocationMini is the campus summary embedded in a profile response.
type FacultyLocationMini struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Region string `json:"region"`
}

// OnboardingStatusResponse tells the frontend how many people the first-time
// user must follow (capped at the followable pool) and how many they already do.
type UsernameChangeItem struct {
	OldUsername string              `json:"old_username"`
	NewUsername string              `json:"new_username"`
	ChangedAt   utils.FormattedTime `json:"changed_at"`
}

type UsernameHistoryResponse struct {
	Count   int                  `json:"count"`
	Changes []UsernameChangeItem `json:"changes"`
}

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
