package usersettingsdto

type UserSettingsResponse struct {
	UserID    uint   `json:"user_id"`
	IsPrivate bool   `json:"is_private"`
	UserTheme string `json:"user_theme"`
}

type UpdateUserSettingsRequest struct {
	IsPrivate bool `json:"is_private"`
}

type UpdateUserThemeRequest struct {
	UserTheme string `json:"user_theme"`
}
