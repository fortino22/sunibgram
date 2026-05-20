package dto

import "github.com/sunibgram/dto/utils"

type MentionUserResponse struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	IsPremium   bool   `json:"is_premium"`
	UserProfile string `json:"user_profile"`
	IsFollowing bool   `json:"is_following"`
}

type CreateThreadRequest struct {
	Content  string  `json:"content"`
	ParentID *uint   `json:"parent_id"`
	EventID  *string `json:"event_id"`
}

type UpdateThreadRequest struct {
	Content string  `json:"content"`
	EventID *string `json:"event_id"`
}

type ThreadResponse struct {
	ThreadID      uint                  `json:"thread_id"`
	UserID        uint                  `json:"user_id"`
	Username      string                `json:"username"`
	Content       string                `json:"content"`
	ParentID      *uint                 `json:"parent_id,omitempty"`
	RootID        *uint                 `json:"root_id,omitempty"`
	EventID       *string               `json:"event_id,omitempty"`
	EventName     *string               `json:"event_name,omitempty"`
	Media         []ThreadMediaResponse `json:"media"`
	LikeCount     int                   `json:"like_count"`
	RepostCount   int                   `json:"repost_count"`
	BookmarkCount int                   `json:"bookmark_count"`
	CommentCount  int                   `json:"comment_count"`
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
	Parent *ThreadResponse `json:"parent,omitempty"`
	Reply  ThreadResponse  `json:"reply"`
}

type FeedAuthorFaculty struct {
	FacultyID   uint   `json:"faculty_id"`
	FacultyName string `json:"faculty_name"`
}

type FeedAuthor struct {
	ID        uint               `json:"id"`
	Name      string             `json:"name"`
	Username  string             `json:"username"`
	IsPremium bool               `json:"is_premium"`
	AvatarURL string             `json:"avatar_url,omitempty"`
	Faculty   *FeedAuthorFaculty `json:"faculty,omitempty"`
}

type FeedStats struct {
	Likes     int `json:"likes"`
	Reposts   int `json:"reposts"`
	Bookmarks int `json:"bookmarks"`
	Comments  int `json:"comments"`
}

type FeedParentContext struct {
	ID        uint                  `json:"id"`
	ParentID  *uint                 `json:"parent_id,omitempty"`
	RootID    *uint                 `json:"root_id,omitempty"`
	EventID   *string               `json:"event_id,omitempty"`
	EventName *string               `json:"event_name,omitempty"`
	Community *FeedCommunity        `json:"community,omitempty"`
	Author    FeedAuthor            `json:"author"`
	Content   string                `json:"content"`
	Media     []ThreadMediaResponse `json:"media"`
	Stats     FeedStats             `json:"stats"`
	CreatedAt utils.FormattedTime   `json:"created_at"`
	UpdatedAt utils.FormattedTime   `json:"updated_at"`
}

type FeedCommunity struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type FeedItemResponse struct {
	Type               string                `json:"type"`
	ParentContext      *FeedParentContext    `json:"parent_context,omitempty"`
	ID                 uint                  `json:"id"`
	ParentID           *uint                 `json:"parent_id,omitempty"`
	RootID             *uint                 `json:"root_id,omitempty"`
	EventID            *string               `json:"event_id,omitempty"`
	EventName          *string               `json:"event_name,omitempty"`
	Community          *FeedCommunity        `json:"community,omitempty"`
	Content            string                `json:"content"`
	CreatedAt          utils.FormattedTime   `json:"created_at"`
	UpdatedAt          utils.FormattedTime   `json:"updated_at"`
	TimeAgo            string                `json:"time_ago"`
	ActionAt           *utils.FormattedTime  `json:"action_at,omitempty"`
	RepostedAt         *utils.FormattedTime  `json:"reposted_at,omitempty"`
	RepostedByID       *uint                 `json:"reposted_by_id,omitempty"`
	RepostedByUsername string                `json:"reposted_by_username,omitempty"`
	RepostedByName     string                `json:"reposted_by_name,omitempty"`
	OriginalThreadID   *uint                 `json:"original_thread_id,omitempty"`
	Author             FeedAuthor            `json:"author"`
	Stats              FeedStats             `json:"stats"`
	Media              []ThreadMediaResponse `json:"media"`
	IsLiked            *bool                 `json:"is_liked,omitempty"`
	IsBookmarked       *bool                 `json:"is_bookmarked,omitempty"`
	IsReposted         *bool                 `json:"is_reposted,omitempty"`
	IsPinned           *bool                 `json:"is_pin,omitempty"`
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

type MediaOnlyItem struct {
	ThreadID uint                  `json:"thread_id"`
	Media    []ThreadMediaResponse `json:"media"`
}

type PaginatedMediaOnlyResponse struct {
	Data []MediaOnlyItem `json:"data"`
	Meta FeedMeta        `json:"meta"`
}

type CommunityThreadFeedItemResponse struct {
	FeedItemResponse
}

type PaginatedCommunityThreadResponse struct {
	Data []CommunityThreadFeedItemResponse `json:"data"`
	Meta FeedMeta                          `json:"meta"`
}

type CommunityThreadMediaResponse struct {
	ThreadID uint                  `json:"thread_id"`
	Media    []ThreadMediaResponse `json:"media"`
}
