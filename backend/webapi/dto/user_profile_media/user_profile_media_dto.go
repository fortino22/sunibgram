package userprofilemediadto

import utils "github.com/sunibgram/webapi/dto/_shared"

type UserMediaProfileResponse struct {
	ProfileID uint                `json:"profile_id"`
	UserID    uint                `json:"user_id"`
	AvatarURL string              `json:"avatar_url"`
	BannerURL string              `json:"banner_url"`
	Bio       string              `json:"bio"`
	UpdatedAt utils.FormattedTime `json:"updated_at"`
}
