package eventreviewdto

import utils "github.com/sunibgram/webapi/dto/_shared"

type DeclineEventRequest struct {
	Reason string `json:"reason" validate:"required"`
}

type RevisionRequest struct {
	Note string `json:"note" validate:"required"`
}

type EventReviewNoteResponse struct {
	RevisionNumber int                 `json:"revision_number"`
	Notes          string              `json:"notes"`
	CreatedAt      utils.FormattedTime `json:"created_at"`
}

type EventAdminReviewNoteResponse struct {
	RevisionNumber int                 `json:"revision_number"`
	Notes          string              `json:"notes"`
	AdminName      string              `json:"admin_name"`
	CreatedAt      utils.FormattedTime `json:"created_at"`
}
