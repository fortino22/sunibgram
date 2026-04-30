package dto

import "github.com/sunibgram/dto/utils"

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
	IsRepost      bool                  `json:"is_repost"`
	RepostedAt    *utils.FormattedTime  `json:"reposted_at,omitempty"`
	CreatedAt     utils.FormattedTime   `json:"created_at"`
	UpdatedAt     utils.FormattedTime   `json:"updated_at"`
}

type PaginatedThreadsResponse struct {
	Data  []ThreadResponse `json:"data"`
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
}

type ReplyResponse struct {
	Parent *ThreadResponse `json:"parent"`
	Reply  ThreadResponse  `json:"reply"`
}

type FeedAuthor struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

type FeedStats struct {
	Likes     int `json:"likes"`
	Reposts   int `json:"reposts"`
	Bookmarks int `json:"bookmarks"`
}

type FeedParentContext struct {
	ID        uint                  `json:"id"`
	ParentID  *uint                 `json:"parent_id"`
	RootID    *uint                 `json:"root_id"`
	Author    FeedAuthor            `json:"author"`
	Content   string                `json:"content"`
	Media     []ThreadMediaResponse `json:"media"`
	Stats     FeedStats             `json:"stats"`
	CreatedAt utils.FormattedTime   `json:"created_at"`
	UpdatedAt utils.FormattedTime   `json:"updated_at"`
}

type FeedItemResponse struct {
	Type          string                `json:"type"`
	ParentContext *FeedParentContext     `json:"parent_context,omitempty"`
	ID            uint                  `json:"id"`
	ParentID      *uint                 `json:"parent_id"`
	RootID        *uint                 `json:"root_id"`
	Content       string                `json:"content"`
	CreatedAt     utils.FormattedTime   `json:"created_at"`
	UpdatedAt     utils.FormattedTime   `json:"updated_at"`
	RepostedAt    *utils.FormattedTime  `json:"reposted_at,omitempty"`
	ActionAt      *utils.FormattedTime  `json:"action_at,omitempty"`
	Author        FeedAuthor            `json:"author"`
	Stats         FeedStats             `json:"stats"`
	Media         []ThreadMediaResponse `json:"media"`
}

type FeedMeta struct {
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
}

type PaginatedFeedResponse struct {
	Data []FeedItemResponse `json:"data"`
	Meta FeedMeta           `json:"meta"`
}
