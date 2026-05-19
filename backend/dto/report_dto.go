package dto

import "github.com/sunibgram/dto/utils"

type CreateReportRequest struct {
	ReportedUserID *uint  `json:"reported_user_id"`
	ReportedPostID *uint  `json:"reported_post_id"`
	ViolationID    string `json:"violation_id"`
	Explanation    string `json:"explanation"`
}

type UpdateReportStatusRequest struct {
	Status string `json:"status"`
}

type ReportResponse struct {
	ID             string              `json:"id"`
	ReporterID     uint                `json:"reporter_id"`
	ReportedUserID *uint               `json:"reported_user_id"`
	ReportedPostID *uint               `json:"reported_post_id"`
	ViolationID    string              `json:"violation_id"`
	Violation      *ViolationResponse  `json:"violation,omitempty"`
	ReporterUser   *ReportUserResponse `json:"reporter_user,omitempty"`
	ReportedUser   *ReportUserResponse `json:"reported_user,omitempty"`
	Explanation    string              `json:"explanation"`
	Status         string              `json:"status"`
	CreatedAt      utils.FormattedTime `json:"created_at"`
	UpdatedAt      utils.FormattedTime `json:"updated_at"`
}

type ReportUserResponse struct {
	UserID     uint                   `json:"user_id"`
	Username   string                 `json:"username"`
	UserAvatar string                 `json:"user_avatar"`
	IsPremium  bool                   `json:"is_premium"`
	Faculty    *ReportFacultyResponse `json:"faculty,omitempty"`
}

type ReportFacultyResponse struct {
	FacultyID   uint   `json:"faculty_id"`
	FacultyName string `json:"faculty_name"`
}
