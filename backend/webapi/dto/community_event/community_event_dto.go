package communityeventdto

import (
	utils "github.com/sunibgram/webapi/dto/_shared"
	eventdto "github.com/sunibgram/webapi/dto/event"
	threaddto "github.com/sunibgram/webapi/dto/thread"
)

// EventTime reuses the global event time marshaler ("2006-01-02 15:04:05").
type EventTime = eventdto.EventTime

// AuthorMini is a compact user summary used for creators, invitees, and comment
// authors.
type AuthorMini struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type ReactionSummaryItem struct {
	Emoji string `json:"emoji"`
	Count int    `json:"count"`
}

type CommentResponse struct {
	ID        uint                `json:"id"`
	Author    AuthorMini          `json:"author"`
	Content   string              `json:"content"`
	IsPinned  bool                `json:"is_pinned"`
	CreatedAt utils.FormattedTime `json:"created_at"`
}

type CommunityEventResponse struct {
	ID              uint                  `json:"id"`
	CommunityID     uint                  `json:"community_id"`
	CreatedBy       AuthorMini            `json:"created_by"`
	Title           string                `json:"title"`
	Description     string                `json:"description"`
	Location        string                `json:"location"`
	StartTime       EventTime             `json:"start_time"`
	EndTime         EventTime             `json:"end_time"`
	BannerURL       string                `json:"banner_url,omitempty"`
	Invitees        []AuthorMini          `json:"invitees"`
	ReactionSummary []ReactionSummaryItem `json:"reaction_summary"`
	MyReaction      string                `json:"my_reaction,omitempty"`
	PinnedComment   *CommentResponse      `json:"pinned_comment,omitempty"`
	CommentCount    int                   `json:"comment_count"`
	CanComment      bool                  `json:"can_comment"`
	IsAdmin         bool                  `json:"is_admin"`
	CreatedAt       utils.FormattedTime   `json:"created_at"`
	UpdatedAt       utils.FormattedTime   `json:"updated_at"`
}

type CommunityEventListResponse struct {
	ID           uint                `json:"id"`
	CommunityID  uint                `json:"community_id"`
	Title        string              `json:"title"`
	Description  string              `json:"description"`
	Location     string              `json:"location"`
	StartTime    EventTime           `json:"start_time"`
	EndTime      EventTime           `json:"end_time"`
	BannerURL    string              `json:"banner_url,omitempty"`
	CommentCount int                 `json:"comment_count"`
	CreatedAt    utils.FormattedTime `json:"created_at"`
}

type PaginatedCommunityEventResponse struct {
	Data []CommunityEventListResponse `json:"data"`
	Meta threaddto.FeedMeta           `json:"meta"`
}

type PaginatedCommentResponse struct {
	Data            []CommentResponse  `json:"data"`
	PinnedCommentID *uint              `json:"pinned_comment_id,omitempty"`
	IsAdmin         bool               `json:"is_admin"`
	Meta            threaddto.FeedMeta `json:"meta"`
}

type AddCommentRequest struct {
	Content string `json:"content"`
}

type SetReactionRequest struct {
	Emoji string `json:"emoji"`
}
