package dto

import "time"

type CreateThreadRequest struct {
	Content  string `json:"content"`
	ParentID *uint  `json:"parent_id"`
}

type UpdateThreadRequest struct {
	Content string `json:"content"`
}

type ThreadResponse struct {
	ThreadID      uint                  `json:"thread_id"`
	UserID        uint                  `json:"user_id"`
	Username      string                `json:"username"`
	Content       string                `json:"content"`
	ParentID      *uint                 `json:"parent_id"`
	RootID        *uint                 `json:"root_id"`
	Media         []ThreadMediaResponse `json:"media"`
	LikeCount     int                   `json:"like_count"`
	RepostCount   int                   `json:"repost_count"`
	BookmarkCount int                   `json:"bookmark_count"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
}

type PaginatedThreadsResponse struct {
	Data  []ThreadResponse `json:"data"`
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
}
