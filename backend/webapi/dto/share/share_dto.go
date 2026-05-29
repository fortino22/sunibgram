package sharedto

type ShareViaChatRequest struct {
	UserIDs []uint  `json:"user_ids"`
	Message *string `json:"message,omitempty"`
}

type CopyShareResponse struct {
	Message    string `json:"message"`
	ShareCount int    `json:"share_count"`
}

type ChatShareResponse struct {
	Message        string `json:"message"`
	RecipientCount int    `json:"recipient_count"`
}

type ShareUserResponse struct {
	UserID    uint    `json:"user_id"`
	Name      string  `json:"name"`
	Username  string  `json:"username"`
	IsPremium bool    `json:"is_premium"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

type SearchShareUsersResponse struct {
	Data []ShareUserResponse `json:"data"`
}
