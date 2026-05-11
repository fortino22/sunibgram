package dto

import (
	dtoutils "github.com/sunibgram/dto/utils"
)

type CreateNotificationRequest struct {
	Type            string `json:"type"`
	Message         string `json:"message"`
	SenderUserID    uint   `json:"sender_user_id"`
	RecipientUserID uint   `json:"recipient_user_id"`
	ReferenceID     *uint  `json:"reference_id"`
}

type NotificationFacultyResponse struct {
	FacultyID   uint   `json:"faculty_id"`
	FacultyName string `json:"faculty_name"`
	TotalUsers  int    `json:"total_users"`
}

type NotificationSenderResponse struct {
	UserID    uint                         `json:"user_id"`
	Username  string                       `json:"username"`
	Name      string                       `json:"name"`
	AvatarURL string                       `json:"avatar_url"`
	IsPremium bool                         `json:"is_premium"`
	Faculty   *NotificationFacultyResponse `json:"faculty,omitempty"`
}

type NotificationListResponse struct {
	ID        uint                       `json:"id"`
	Type      string                     `json:"type"`
	Message   string                     `json:"message"`
	CreatedAt dtoutils.FormattedTime     `json:"created_at"`
	Actor     NotificationSenderResponse `json:"actor"`
}

type NotificationListMeta struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type PaginatedNotificationResponse struct {
	Data []NotificationListResponse `json:"data"`
	Meta NotificationListMeta       `json:"meta"`
}
