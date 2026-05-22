package usersettingsdto

type UserSettingsResponse struct {
	UserID    uint `json:"user_id"`
	IsPrivate bool `json:"is_private"`
}

type UpdateUserSettingsRequest struct {
	IsPrivate bool `json:"is_private"`
}
