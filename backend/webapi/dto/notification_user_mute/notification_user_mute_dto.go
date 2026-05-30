package notificationusermutedto

import (
	threaddto "github.com/sunibgram/webapi/dto/thread"
)

type MutedUserResponse struct {
	UserID    uint   `json:"user_id"`
	Name      string `json:"name"`
	UserName  string `json:"username"`
	IsPremium bool   `json:"is_premium"`
	AvatarURL string `json:"avatar_url"`
}

type PaginatedMutedUsersResponse struct {
	Data []MutedUserResponse `json:"data"`
	Meta threaddto.FeedMeta  `json:"meta"`
}
