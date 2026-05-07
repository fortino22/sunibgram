package dto

type FollowUserResponse struct {
	UserID    uint   `json:"user_id"`
	Name      string `json:"name"`
	UserName  string `json:"username"`
	IsPremium bool   `json:"is_premium"`
	AvatarURL string `json:"avatar_url"`
}

type PaginatedFollowUsersResponse struct {
	Data []FollowUserResponse `json:"data"`
	Meta FeedMeta             `json:"meta"`
}
