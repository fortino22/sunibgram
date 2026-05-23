package feedbackdto

import utils "github.com/sunibgram/webapi/dto/_shared"

type CreateFeedbackRequest struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

type UpdateStatusRequest struct {
	Status string `json:"status"`
}

type FeedbackAuthor struct {
	UserID     uint   `json:"user_id"`
	Username   string `json:"username"`
	Name       string `json:"name"`
	UserAvatar string `json:"user_avatar"`
	IsPremium  bool   `json:"is_premium"`
}

type FeedbackResponse struct {
	ID        string              `json:"id"`
	Type      string              `json:"type"`
	Title     string              `json:"title"`
	Body      string              `json:"body"`
	Status    string              `json:"status"`
	Author    FeedbackAuthor      `json:"author"`
	Likes     int64               `json:"likes"`
	Dislikes  int64               `json:"dislikes"`
	IsVote    *string             `json:"is_vote"`
	CreatedAt utils.FormattedTime `json:"created_at"`
	UpdatedAt utils.FormattedTime `json:"updated_at"`
}

type FeedbackListMeta struct {
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
}

type PaginatedFeedbackResponse struct {
	Data []FeedbackResponse `json:"data"`
	Meta FeedbackListMeta   `json:"meta"`
}
