package dto

import "github.com/sunibgram/dto/utils"

type DeclineEventRequest struct {
	Reason string `json:"reason" validate:"required"`
}

type RevisionRequest struct {
	Note string `json:"note" validate:"required"`
}

type EventReviewResponse struct {
	ID        string              `json:"id"`
	EventID   string              `json:"event_id"`
	AdminID   uint                `json:"admin_id"`
	Action    string              `json:"action"`
	Note      string              `json:"note"`
	CreatedAt utils.FormattedTime `json:"created_at"`
}
